package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/t-kuni/cqrs-example/di"
	"github.com/t-kuni/cqrs-example/domain/infrastructure/db"
	"github.com/t-kuni/cqrs-example/domain/model"
	"go.uber.org/fx"
)

func main() {
	godotenv.Load(filepath.Join(".env"))

	ctx := context.Background()
	app := di.NewApp(fx.Invoke(func(conn db.IConnector) {
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

		_, err = db.Exec("SET sql_log_bin=0")
		if err != nil {
			panic(err)
		}

		// 各テーブルをTRUNCATEする
		tables := []string{"products", "categories", "tenants", "users"}
		for _, table := range tables {
			_, err := db.Exec("TRUNCATE TABLE " + table)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Truncated table: %s\n", table)
		}

		// CSV ファイルの出力先ディレクトリ
		tmpDir := "/tmp"

		// 1. Users用のCSVを生成
		fmt.Println("Creating users CSV...")
		userIDs, err := createUsersCSV(tmpDir, 200)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Created users.csv with %d records\n", len(userIDs))

		// 2. Tenants用のCSVを生成
		fmt.Println("Creating tenants CSV...")
		tenantIDs, err := createTenantsCSV(tmpDir, userIDs, 1000)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Created tenants.csv with %d records\n", len(tenantIDs))

		// 3. Categories用のCSVを生成
		fmt.Println("Creating categories CSV...")
		categoryIDs, err := createCategoriesCSV(tmpDir, 50)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Created categories.csv with %d records\n", len(categoryIDs))

		// 4. Products用のCSVを生成
		fmt.Println("Creating products CSV...")
		productsCount := int32(1000000)
		err = createProductsCSV(tmpDir, tenantIDs, categoryIDs, productsCount)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Created products.csv with %d records\n", productsCount)

		// 5. LOAD DATA LOCAL INFILE を実行してデータを投入
		fmt.Println("Loading data from CSV files...")

		// Users
		_, err = db.Exec("BEGIN")
		if err != nil {
			panic(err)
		}
		usersCSVPath := filepath.Join(tmpDir, "users.csv")
		_, err = db.Exec(fmt.Sprintf("LOAD DATA LOCAL INFILE '%s' INTO TABLE users FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\\n' (id, name)", usersCSVPath))
		if err != nil {
			panic(fmt.Errorf("failed to load users.csv: %w", err))
		}
		_, err = db.Exec("COMMIT")
		if err != nil {
			panic(err)
		}
		fmt.Println("  Loaded users")

		// Tenants
		_, err = db.Exec("BEGIN")
		if err != nil {
			panic(err)
		}
		tenantsCSVPath := filepath.Join(tmpDir, "tenants.csv")
		_, err = db.Exec(fmt.Sprintf("LOAD DATA LOCAL INFILE '%s' INTO TABLE tenants FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\\n' (id, owner_id, name)", tenantsCSVPath))
		if err != nil {
			panic(fmt.Errorf("failed to load tenants.csv: %w", err))
		}
		_, err = db.Exec("COMMIT")
		if err != nil {
			panic(err)
		}
		fmt.Println("  Loaded tenants")

		// Categories
		_, err = db.Exec("BEGIN")
		if err != nil {
			panic(err)
		}
		categoriesCSVPath := filepath.Join(tmpDir, "categories.csv")
		_, err = db.Exec(fmt.Sprintf("LOAD DATA LOCAL INFILE '%s' INTO TABLE categories FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\\n' (id, name)", categoriesCSVPath))
		if err != nil {
			panic(fmt.Errorf("failed to load categories.csv: %w", err))
		}
		_, err = db.Exec("COMMIT")
		if err != nil {
			panic(err)
		}
		fmt.Println("  Loaded categories")

		// Products
		_, err = db.Exec("BEGIN")
		if err != nil {
			panic(err)
		}
		productsCSVPath := filepath.Join(tmpDir, "products.csv")
		_, err = db.Exec(fmt.Sprintf("LOAD DATA LOCAL INFILE '%s' INTO TABLE products FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\\n' (id, tenant_id, category_id, name, price, properties, listed_at)", productsCSVPath))
		if err != nil {
			panic(fmt.Errorf("failed to load products.csv: %w", err))
		}
		_, err = db.Exec("COMMIT")
		if err != nil {
			panic(err)
		}
		fmt.Println("  Loaded products")

		fmt.Println("Seeding successfully!")
	}))

	defer app.Stop(ctx)
	err := app.Start(ctx)
	if err != nil {
		panic(err)
	}
}

// createUsersCSV creates users.csv and returns user IDs
func createUsersCSV(dir string, count int32) ([]uuid.UUID, error) {
	filePath := filepath.Join(dir, "users.csv")
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	ids := make([]uuid.UUID, count)
	for i := int32(0); i < count; i++ {
		id := uuid.New()
		ids[i] = id
		name := fmt.Sprintf("ユーザ%d", i+1)

		record := []string{id.String(), name}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	return ids, nil
}

// createTenantsCSV creates tenants.csv and returns tenant IDs
func createTenantsCSV(dir string, userIDs []uuid.UUID, count int32) ([]uuid.UUID, error) {
	filePath := filepath.Join(dir, "tenants.csv")
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	ids := make([]uuid.UUID, count)
	usersPerTenant := int32(5) // 1000 tenants / 200 users = 5 tenants per user

	for i := int32(0); i < count; i++ {
		id := uuid.New()
		ids[i] = id
		name := fmt.Sprintf("テナント%d", i+1)

		userIndex := i / usersPerTenant
		if int32(userIndex) >= int32(len(userIDs)) {
			userIndex = int32(len(userIDs)) - 1
		}
		ownerID := userIDs[userIndex]

		record := []string{id.String(), ownerID.String(), name}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	return ids, nil
}

// createCategoriesCSV creates categories.csv and returns category IDs
func createCategoriesCSV(dir string, count int32) ([]uuid.UUID, error) {
	filePath := filepath.Join(dir, "categories.csv")
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	ids := make([]uuid.UUID, count)
	for i := int32(0); i < count; i++ {
		id := uuid.New()
		ids[i] = id
		name := fmt.Sprintf("カテゴリ%d", i+1)

		record := []string{id.String(), name}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	return ids, nil
}

// createProductsCSV creates products.csv
func createProductsCSV(dir string, tenantIDs, categoryIDs []uuid.UUID, count int32) error {
	filePath := filepath.Join(dir, "products.csv")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	numTenants := int32(len(tenantIDs))
	numCategories := int32(len(categoryIDs))
	now := time.Now()
	oneYearAgo := now.AddDate(-1, 0, 0)
	yearInSeconds := int64(now.Sub(oneYearAgo).Seconds())

	sizes := []string{"S", "M", "L"}
	colors := []string{"red", "green", "blue"}

	for i := int32(0); i < count; i++ {
		id := uuid.New()
		name := fmt.Sprintf("商品%d", i+1)
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

		// JSON化
		propertiesJSON, err := json.Marshal(properties)
		if err != nil {
			return err
		}

		// listed_at: random time within the past year
		randomSeconds := rand.Int63n(yearInSeconds)
		listedAt := oneYearAgo.Add(time.Duration(randomSeconds) * time.Second)
		listedAtStr := listedAt.Format("2006-01-02 15:04:05")

		// Distribute products evenly across tenants and categories
		tenantID := tenantIDs[i%numTenants]
		categoryID := categoryIDs[i%numCategories]

		record := []string{
			id.String(),
			tenantID.String(),
			categoryID.String(),
			name,
			fmt.Sprintf("%d", price),
			string(propertiesJSON),
			listedAtStr,
		}
		if err := writer.Write(record); err != nil {
			return err
		}

		// Progress display
		if (i+1)%10000 == 0 || i+1 == count {
			fmt.Printf("  Progress: %d/%d products generated\n", i+1, count)
		}
	}

	return nil
}
