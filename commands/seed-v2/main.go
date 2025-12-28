package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
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

		// productsテーブルのキー・インデックス情報を取得
		fmt.Println("Backing up products table indexes and foreign keys...")
		indexes, err := getProductsIndexes(db)
		if err != nil {
			panic(fmt.Errorf("failed to get products indexes: %w", err))
		}
		fmt.Printf("  Found %d index(es)\n", len(indexes))

		foreignKeys, err := getProductsForeignKeys(db)
		if err != nil {
			panic(fmt.Errorf("failed to get products foreign keys: %w", err))
		}
		fmt.Printf("  Found %d foreign key(s)\n", len(foreignKeys))

		// productsテーブルのキー・インデックスを削除
		fmt.Println("Dropping products table indexes and foreign keys...")
		err = dropProductsForeignKeys(db, foreignKeys)
		if err != nil {
			panic(fmt.Errorf("failed to drop products foreign keys: %w", err))
		}

		err = dropProductsIndexes(db, indexes)
		if err != nil {
			panic(fmt.Errorf("failed to drop products indexes: %w", err))
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

		// productsテーブルのキー・インデックスを復元
		fmt.Println("Restoring products table indexes and foreign keys...")
		err = restoreProductsIndexes(db, indexes)
		if err != nil {
			panic(fmt.Errorf("failed to restore products indexes: %w", err))
		}

		err = restoreProductsForeignKeys(db, foreignKeys)
		if err != nil {
			panic(fmt.Errorf("failed to restore products foreign keys: %w", err))
		}

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

// IndexInfo holds information about an index
type IndexInfo struct {
	IndexName  string
	ColumnName string
	NonUnique  bool
	SeqInIndex int64
}

// ForeignKeyInfo holds information about a foreign key
type ForeignKeyInfo struct {
	ConstraintName       string
	ColumnName           string
	ReferencedTableName  string
	ReferencedColumnName string
}

// getProductsIndexes retrieves index information for the products table
func getProductsIndexes(db *sql.DB) ([]IndexInfo, error) {
	query := `
		SELECT DISTINCT INDEX_NAME, COLUMN_NAME, NON_UNIQUE, SEQ_IN_INDEX
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'products'
		  AND INDEX_NAME != 'PRIMARY'
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var idx IndexInfo
		var nonUnique int64
		if err := rows.Scan(&idx.IndexName, &idx.ColumnName, &nonUnique, &idx.SeqInIndex); err != nil {
			return nil, fmt.Errorf("failed to scan index row: %w", err)
		}
		idx.NonUnique = nonUnique == 1
		indexes = append(indexes, idx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating index rows: %w", err)
	}

	return indexes, nil
}

// getProductsForeignKeys retrieves foreign key information for the products table
func getProductsForeignKeys(db *sql.DB) ([]ForeignKeyInfo, error) {
	query := `
		SELECT CONSTRAINT_NAME, COLUMN_NAME, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'products'
		  AND REFERENCED_TABLE_NAME IS NOT NULL
		ORDER BY CONSTRAINT_NAME
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	var fks []ForeignKeyInfo
	for rows.Next() {
		var fk ForeignKeyInfo
		if err := rows.Scan(&fk.ConstraintName, &fk.ColumnName, &fk.ReferencedTableName, &fk.ReferencedColumnName); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key row: %w", err)
		}
		fks = append(fks, fk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating foreign key rows: %w", err)
	}

	return fks, nil
}

// dropProductsIndexes drops indexes from the products table
func dropProductsIndexes(db *sql.DB, indexes []IndexInfo) error {
	// Group indexes by name to handle composite indexes
	indexMap := make(map[string]bool)
	for _, idx := range indexes {
		indexMap[idx.IndexName] = true
	}

	for indexName := range indexMap {
		query := fmt.Sprintf("ALTER TABLE products DROP INDEX `%s`", indexName)
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to drop index %s: %w", indexName, err)
		}
		fmt.Printf("  Dropped index: %s\n", indexName)
	}

	return nil
}

// dropProductsForeignKeys drops foreign keys from the products table
func dropProductsForeignKeys(db *sql.DB, fks []ForeignKeyInfo) error {
	// Group foreign keys by constraint name
	fkMap := make(map[string]bool)
	for _, fk := range fks {
		fkMap[fk.ConstraintName] = true
	}

	for constraintName := range fkMap {
		query := fmt.Sprintf("ALTER TABLE products DROP FOREIGN KEY `%s`", constraintName)
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to drop foreign key %s: %w", constraintName, err)
		}
		fmt.Printf("  Dropped foreign key: %s\n", constraintName)
	}

	return nil
}

// restoreProductsIndexes restores indexes to the products table
func restoreProductsIndexes(db *sql.DB, indexes []IndexInfo) error {
	// Group indexes by name to handle composite indexes
	indexMap := make(map[string][]IndexInfo)
	for _, idx := range indexes {
		indexMap[idx.IndexName] = append(indexMap[idx.IndexName], idx)
	}

	for indexName, idxList := range indexMap {
		// Sort by SEQ_IN_INDEX to maintain column order
		// (already sorted by query ORDER BY clause)

		// Build column list
		var columns []string
		for _, idx := range idxList {
			columns = append(columns, fmt.Sprintf("`%s`", idx.ColumnName))
		}

		// Determine if UNIQUE or not
		unique := ""
		if !idxList[0].NonUnique {
			unique = "UNIQUE "
		}

		query := fmt.Sprintf("ALTER TABLE products ADD %sINDEX `%s` (%s)", unique, indexName, strings.Join(columns, ", "))
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to restore index %s: %w", indexName, err)
		}
		fmt.Printf("  Restored index: %s\n", indexName)
	}

	return nil
}

// restoreProductsForeignKeys restores foreign keys to the products table
func restoreProductsForeignKeys(db *sql.DB, fks []ForeignKeyInfo) error {
	// Group foreign keys by constraint name to handle composite foreign keys
	fkMap := make(map[string][]ForeignKeyInfo)
	for _, fk := range fks {
		fkMap[fk.ConstraintName] = append(fkMap[fk.ConstraintName], fk)
	}

	for constraintName, fkList := range fkMap {
		// Build column list and referenced column list
		var columns []string
		var refColumns []string
		for _, fk := range fkList {
			columns = append(columns, fmt.Sprintf("`%s`", fk.ColumnName))
			refColumns = append(refColumns, fmt.Sprintf("`%s`", fk.ReferencedColumnName))
		}

		// All FKs in the list should have the same referenced table
		refTable := fkList[0].ReferencedTableName

		query := fmt.Sprintf(
			"ALTER TABLE products ADD CONSTRAINT `%s` FOREIGN KEY (%s) REFERENCES `%s` (%s)",
			constraintName,
			strings.Join(columns, ", "),
			refTable,
			strings.Join(refColumns, ", "),
		)
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to restore foreign key %s: %w", constraintName, err)
		}
		fmt.Printf("  Restored foreign key: %s\n", constraintName)
	}

	return nil
}
