package base_swagger_service

import (
	"context"
	"crypto/tls"
	"net/http"

	openapiErrors "github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

type baseServiceImpl struct {
	name                  string
	version               string
	metrics               *prometheus.Registry
	shutdownCtx           context.Context
	shutdownCtxCancelFunc func()

	statReqCount     *prometheus.CounterVec
	statReqDurations *prometheus.HistogramVec

	middlewareLogger MiddlewareLogger
}

// NewBaseService создает базовый сервис.
func NewBaseService(name, version string) (BaseSwaggerService, error) {
	srv := &baseServiceImpl{
		name:    name,
		version: version,
	}

	srv.shutdownCtx, srv.shutdownCtxCancelFunc = context.WithCancel(context.Background())

	// инициализация Profiler
	if err := srv.initProfiler(); err != nil {
		return nil, err
	}

	// инициализация метрик
	srv.initMetrics()

	srv.middlewareLogger = NewDebugLogger()

	return srv, nil
}

func (srv *baseServiceImpl) Name() string {
	return srv.name
}

func (srv *baseServiceImpl) Version() string {
	return srv.version
}

func (srv *baseServiceImpl) Metrics() Prometheus {
	return srv.metrics
}

func (srv *baseServiceImpl) ShutdownContext() context.Context {
	return srv.shutdownCtx
}

func (srv *baseServiceImpl) ShutdownCallback(implShutdownCallback func()) {
	srv.shutdownCtxCancelFunc()

	// реализация shutdown конкретного сервиса
	implShutdownCallback()
}

func (srv *baseServiceImpl) LogCallback(str string, args ...interface{}) {
	logrus.Infof(str, args...)
}

func (srv *baseServiceImpl) ServeError(rw http.ResponseWriter, r *http.Request, err error) {
	openapiErrors.ServeError(rw, r, err)
}

func (srv *baseServiceImpl) ConfigureTLS(tlsConfig *tls.Config) {
}

func (srv *baseServiceImpl) ConfigureServer(scheme string, addr string) {
}

func (srv *baseServiceImpl) SetupMiddlewares(ctx *middleware.Context, handler http.Handler) http.Handler {
	return srv.recoverMiddleware(
		srv.profileMiddleware(
			ctx,
			srv.metricsMiddleware(
				ctx,
				srv.loggerMiddleware(
					ctx,
					handler,
				),
			),
		),
	)
}

func (srv *baseServiceImpl) SetupMiddlewareLogger(logger MiddlewareLogger) {
	srv.middlewareLogger = logger
}

func (srv *baseServiceImpl) initProfiler() error {
	var conf struct {
		NewRelic struct {
			Key     string `envconfig:"optional"`
			Proxy   string `envconfig:"optional"`
			Enabled bool   `envconfig:"default=false"`
		}
	}

	if err := envconfig.InitWithPrefix(&conf, srv.Name()); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (srv *baseServiceImpl) initMetrics() {
	srv.metrics = prometheus.NewRegistry()

	srv.metrics.MustRegister(prometheus.NewGoCollector())
	srv.metrics.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{
		PidFn:     nil,
		Namespace: srv.name,
	}))

	srv.statReqCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: srv.name,
		Name:      "rest_requests_total",
		Help:      "Total number of rest requests",
	}, []string{"method", "code", "handler"})

	srv.statReqDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: srv.name,
		Name:      "rest_request_duration_seconds",
		Help:      "Rest request duration",
		Buckets:   []float64{0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
	}, []string{"handler"})

	srv.metrics.MustRegister(srv.statReqCount, srv.statReqDurations)
}
