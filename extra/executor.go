package extra

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	botdetector "github.com/krakend/krakend-botdetector/v2/gin"
	"github.com/krakend/krakend-ce/v2"
	cmd "github.com/krakend/krakend-cobra/v2"
	jose "github.com/krakend/krakend-jose/v2"
	ginjose "github.com/krakend/krakend-jose/v2/gin"
	lua "github.com/krakend/krakend-lua/v2/router/gin"
	metrics "github.com/krakend/krakend-metrics/v2/gin"
	opencensus "github.com/krakend/krakend-opencensus/v2/router/gin"
	ratelimit "github.com/krakend/krakend-ratelimit/v3/router/gin"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	router "github.com/luraproject/lura/v2/router/gin"
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
