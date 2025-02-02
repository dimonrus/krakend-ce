package extra

import (
	"context"
	"fmt"
	"github.com/krakendio/krakend-ce/v2"
	cmd "github.com/krakendio/krakend-cobra/v2"
	jose "github.com/krakendio/krakend-jose/v2"
	ginjose "github.com/krakendio/krakend-jose/v2/gin"
	lua "github.com/krakendio/krakend-lua/v2/router/gin"
	opencensus "github.com/krakendio/krakend-opencensus/v2/router/gin"
	ratelimit "github.com/krakendio/krakend-ratelimit/v3/router/gin"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/proxy"
	router "github.com/luraproject/lura/v2/router/gin"
	"net/http"

	"github.com/gin-gonic/gin"

	botdetector "github.com/krakendio/krakend-botdetector/v2/gin"
	metrics "github.com/krakendio/krakend-metrics/v2/gin"
	"github.com/luraproject/lura/v2/logging"
)

type IPortError interface {
	GetHTTP() int
}

// ProxyErrorHandler error processing
func ProxyErrorHandler(err error) int {
	if e, ok := err.(IPortError); ok {
		return e.GetHTTP()
	}
	return http.StatusInternalServerError
}

// NewHandlerFactoryExtra returns a HandlerFactory with a rate-limit and a metrics collector middleware injected
func NewHandlerFactoryExtra(logger logging.Logger, metricCollector *metrics.Metrics, rejecter jose.RejecterFactory) router.HandlerFactory {
	handlerFactory := router.CustomErrorEndpointHandler(logger, ProxyErrorHandler)
	handlerFactory = ratelimit.NewRateLimiterMw(logger, handlerFactory)
	handlerFactory = lua.HandlerFactory(logger, handlerFactory)
	handlerFactory = ginjose.HandlerFactory(handlerFactory, logger, rejecter)
	handlerFactory = metricCollector.NewHTTPHandlerFactory(handlerFactory)
	handlerFactory = opencensus.New(handlerFactory)
	handlerFactory = botdetector.New(handlerFactory, logger)

	return func(cfg *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
		logger.Debug(fmt.Sprintf("[ENDPOINT: %s] Building the http handler", cfg.Endpoint))
		return handlerFactory(cfg, p)
	}
}

type handlerFactoryExtra struct{}

func (handlerFactoryExtra) NewHandlerFactory(l logging.Logger, m *metrics.Metrics, r jose.RejecterFactory) router.HandlerFactory {
	return NewHandlerFactoryExtra(l, m, r)
}

// NewExecutorExtra returns an executor for the cmd package. The executor initalizes the entire gateway by
// registering the components and composing a RouterFactory wrapping all the middlewares.
func NewExecutorExtra(ctx context.Context) cmd.Executor {
	eb := new(krakend.ExecutorBuilder)
	eb.HandlerFactory = handlerFactoryExtra{}
	return eb.NewCmdExecutor(ctx)
}
