package di

import (
	"github.com/t-kuni/cqrs-example/domain/service"
	customErrors "github.com/t-kuni/cqrs-example/errors"
	"github.com/t-kuni/cqrs-example/infrastructure/api"
	"github.com/t-kuni/cqrs-example/infrastructure/db"
	"github.com/t-kuni/cqrs-example/infrastructure/system"
	"github.com/t-kuni/cqrs-example/middleware"
	"github.com/t-kuni/cqrs-example/validator"
	"go.uber.org/fx"
)

func NewApp(opts ...fx.Option) *fx.App {
	mergedOpts := []fx.Option{
		//fx.WithLogger(func(log *logger.Logger) fxevent.Logger {
		//	return log
		//}),
		fx.Provide(

			// Validator
			validator.NewCustomValidator,

			// Middleware
			middleware.NewRecover,
			middleware.NewAccessLog,

			// Handler
			// handler.NewGetUsers,

			// Service
			service.NewExampleService,

			// Infrastructure
			db.NewConnector,
			api.NewBinanceApi,
			system.NewTimer,
			system.NewLogger,
			system.NewUuidGenerator,

			// Others
			customErrors.NewCustomServeError,
		),
	}
	mergedOpts = append(mergedOpts, opts...)

	return fx.New(mergedOpts...)
}
