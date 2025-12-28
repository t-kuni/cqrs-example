package main

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/t-kuni/cqrs-example/di"
	"github.com/t-kuni/cqrs-example/domain/infrastructure/db"
	"github.com/t-kuni/cqrs-example/domain/model"
	"github.com/t-kuni/cqrs-example/ent"
	"go.uber.org/fx"
)

func main() {
	godotenv.Load(filepath.Join(".env"))

	ctx := context.Background()
	app := di.NewApp(fx.Invoke(func(conn db.IConnector) {
		client := conn.GetEnt()
		db := conn.GetDB()

		// 乱数生成器の初期化
		rand.Seed(time.Now().UnixNano())

		fmt.Println("Starting seed-v2...")

		// データ投入前のSQL設定
		_, err := db.Exec("SET FOREIGN_KEY_CHECKS=0")
		if err != nil {
			panic(err)
		}
		defer db.Exec("SET FOREIGN_KEY_CHECKS=1")

		_, err = db.Exec("SET AUTOCOMMIT=0")
		if err != nil {
			panic(err)
		}

		_, err = db.Exec("SET sql_log_bin=0")
		if err != nil {
			panic(err)
		}

		// 各テーブルをTRUNCATEする
		tables := []string{"users", "tenants", "categories", "products"}
		for _, table := range tables {
			_, err := db.Exec("TRUNCATE TABLE " + table)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Truncated table: %s\n", table)
		}

		// 1. Usersの作成
		fmt.Println("Creating users...")
		userIDs := createUsers(ctx, client, 200)
		fmt.Printf("Created %d users\n", len(userIDs))

		// 2. Tenantsの作成
		fmt.Println("Creating tenants...")
		tenantIDs := createTenants(ctx, client, userIDs, 1000)
		fmt.Printf("Created %d tenants\n", len(tenantIDs))

		// 3. Categoriesの作成
		fmt.Println("Creating categories...")
		categoryIDs := createCategories(ctx, client, 50)
		fmt.Printf("Created %d categories\n", len(categoryIDs))

		// 4. Productsの作成（バッチ処理）
		fmt.Println("Creating products...")
		totalProducts := int32(1000000)
		batchSize := int32(1000)
		createProducts(ctx, client, tenantIDs, categoryIDs, totalProducts, batchSize)
		fmt.Printf("Created %d products\n", totalProducts)

		fmt.Println("Seeding successfully!")
	}))

	defer app.Stop(ctx)
	err := app.Start(ctx)
	if err != nil {
		panic(err)
	}
}

// createUsers creates user records and returns their IDs
func createUsers(ctx context.Context, client *ent.Client, count int32) []uuid.UUID {
	builders := make([]*ent.UserCreate, count)
	for i := int32(0); i < count; i++ {
		name := fmt.Sprintf("ユーザ%d", i+1)
		builders[i] = client.User.Create().SetName(name)
	}

	users, err := client.User.CreateBulk(builders...).Save(ctx)
	if err != nil {
		panic(err)
	}

	ids := make([]uuid.UUID, len(users))
	for i, user := range users {
		ids[i] = user.ID
	}
	return ids
}

// createTenants creates tenant records and returns their IDs
func createTenants(ctx context.Context, client *ent.Client, userIDs []uuid.UUID, count int32) []uuid.UUID {
	builders := make([]*ent.TenantCreate, count)
	usersPerTenant := int32(5) // 1000 tenants / 200 users = 5 tenants per user

	for i := int32(0); i < count; i++ {
		name := fmt.Sprintf("テナント%d", i+1)
		userIndex := i / usersPerTenant
		if int32(userIndex) >= int32(len(userIDs)) {
			userIndex = int32(len(userIDs)) - 1
		}
		builders[i] = client.Tenant.Create().
			SetName(name).
			SetOwnerID(userIDs[userIndex])
	}

	tenants, err := client.Tenant.CreateBulk(builders...).Save(ctx)
	if err != nil {
		panic(err)
	}

	ids := make([]uuid.UUID, len(tenants))
	for i, tenant := range tenants {
		ids[i] = tenant.ID
	}
	return ids
}

// createCategories creates category records and returns their IDs
func createCategories(ctx context.Context, client *ent.Client, count int32) []uuid.UUID {
	builders := make([]*ent.CategoryCreate, count)
	for i := int32(0); i < count; i++ {
		name := fmt.Sprintf("カテゴリ%d", i+1)
		builders[i] = client.Category.Create().SetName(name)
	}

	categories, err := client.Category.CreateBulk(builders...).Save(ctx)
	if err != nil {
		panic(err)
	}

	ids := make([]uuid.UUID, len(categories))
	for i, category := range categories {
		ids[i] = category.ID
	}
	return ids
}

// createProducts creates product records in batches
func createProducts(ctx context.Context, client *ent.Client, tenantIDs, categoryIDs []uuid.UUID, totalCount, batchSize int32) {
	numTenants := int32(len(tenantIDs))
	numCategories := int32(len(categoryIDs))
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)
	yearInSeconds := int64(now.Sub(oneYearAgo).Seconds())

	sizes := []string{"S", "M", "L"}
	colors := []string{"red", "green", "blue"}

	for i := int32(0); i < totalCount; i += batchSize {
		currentBatchSize := batchSize
		if i+batchSize > totalCount {
			currentBatchSize = totalCount - i
		}

		builders := make([]*ent.ProductCreate, currentBatchSize)
		for j := int32(0); j < currentBatchSize; j++ {
			idx := i + j
			name := fmt.Sprintf("商品%d", idx+1)
			price := int64(rand.Intn(9901) + 100) // 100-10000

			// Properties
			size := sizes[rand.Intn(len(sizes))]
			latitude := fmt.Sprintf("%.6f", 20.43+rand.Float64()*(45.55-20.43))
			longitude := fmt.Sprintf("%.6f", 122.93+rand.Float64()*(153.99-122.93))
			color := colors[rand.Intn(len(colors))]
			properties := &model.ProductProperties{
				Size:      &size,
				Latitude:  &latitude,
				Longitude: &longitude,
				Color:     &color,
			}

			// listed_at: random time within the past year
			randomSeconds := rand.Int63n(yearInSeconds)
			listedAt := oneYearAgo.Add(time.Duration(randomSeconds) * time.Second)

			// Distribute products evenly across tenants and categories
			tenantID := tenantIDs[idx%numTenants]
			categoryID := categoryIDs[idx%numCategories]

			builders[j] = client.Product.Create().
				SetName(name).
				SetPrice(price).
				SetProperties(properties).
				SetListedAt(listedAt).
				SetTenantID(tenantID).
				SetCategoryID(categoryID)
		}

		_, err := client.Product.CreateBulk(builders...).Save(ctx)
		if err != nil {
			panic(err)
		}

		if (i+currentBatchSize)%10000 == 0 || i+currentBatchSize == totalCount {
			fmt.Printf("  Progress: %d/%d products created\n", i+currentBatchSize, totalCount)
		}
	}
}

