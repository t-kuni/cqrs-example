//go:generate go tool mockgen -source=$GOFILE -destination=${GOFILE}_mock.go -package=$GOPACKAGE
package service

import (
	"context"
	"github.com/t-kuni/go-web-api-template/domain/infrastructure/api"
	"github.com/t-kuni/go-web-api-template/domain/infrastructure/db"
)

type ExampleService struct {
	BinanceApi  api.IBinanceApi
	DBConnector db.IConnector
}

type IExampleService interface {
	Exec(ctx context.Context, baseAsset string) (string, error)
}

func NewExampleService(conn db.IConnector, binanceApi api.IBinanceApi) (IExampleService, error) {
	return &ExampleService{
		binanceApi,
		conn,
	}, nil
}

func (s ExampleService) Exec(ctx context.Context, baseAsset string) (string, error) {
	return "", nil
}
