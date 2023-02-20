package base_swagger_service

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

type ServiceImplementation interface {
	ConfigureService() error

	SetupSwaggerHandlers(iapi interface{})

	HealthCheckers() []Checker

	OnShutdown()
}

type BaseSwaggerService interface {
	BaseService
	SwaggerConfigurator
}

type SwaggerConfigurator interface {
	ServeError(rw http.ResponseWriter, r *http.Request, err error)

	LogCallback(str string, args ...interface{})

	ShutdownCallback(implShutdownCallback func())

	ConfigureTLS(tlsConfig *tls.Config)

	ConfigureServer(scheme string, addr string)

	SetupMiddlewares(*middleware.Context, http.Handler) http.Handler

	SetupMiddlewareLogger(logger MiddlewareLogger)
}

type BaseService interface {
	Name() string

	Version() string

	ShutdownContext() context.Context

	Metrics() Prometheus
}

type Prometheus interface {
	prometheus.Registerer
	prometheus.Gatherer
}

type MiddlewareLogger interface {
	IsEnable(ctx context.Context) bool
	LogRequest(ctx context.Context, request string)
	LogResponse(ctx context.Context, response string, kvs ...interface{})
}
