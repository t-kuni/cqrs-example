package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/t-kuni/cqrs-example/di"
	"github.com/t-kuni/cqrs-example/domain/service"
	"go.uber.org/fx"
)

func main() {
	godotenv.Load(filepath.Join(".env"))

	ctx := context.Background()
	app := di.NewApp(fx.Invoke(func(transferService service.IProductTransferService) {
		fmt.Println("Starting product transfer to OpenSearch...")

		err := transferService.TransferAllProducts(ctx)
		if err != nil {
			panic(fmt.Errorf("failed to transfer products: %w", err))
		}

		fmt.Println("Product transfer completed successfully!")
	}))

	defer app.Stop(ctx)
	err := app.Start(ctx)
	if err != nil {
		panic(err)
	}
}
