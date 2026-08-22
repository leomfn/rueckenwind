package main

import (
	"fmt"

	"github.com/leomfn/rueckenwind/internal/handlers"
	"github.com/leomfn/rueckenwind/internal/middleware"
	"github.com/leomfn/rueckenwind/internal/server"
)

func main() {
	rueckenwindServer := server.NewServer(port)

	sameSiteMiddleware := middleware.NewSameSiteMiddleware(domain, debug)

	// Only the data endpoints are rate limited, since they are the ones that
	// reach the upstream APIs.
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimitPerMinute, rateLimitBurst, trustProxyHeaders)

	rootRouter := server.NewRouter("/")
	rootRouter.Handle("GET", "/{$}", handlers.NewGetIndexHandler(fmt.Sprintf("%s/index.html", staticFilesDir)))
	rootRouter.Handle("GET", "/assets/", handlers.NewStaticFilesHandler(fmt.Sprintf("%s/assets", staticFilesDir)), sameSiteMiddleware)
	rootRouter.Handle("GET", "/health", handlers.NewHealthcheckHandler())

	dataRouter := server.NewRouter("/data/")
	dataRouter.Handle("POST", "/weather", handlers.NewWeatherHandler(owmApiKey), sameSiteMiddleware, rateLimitMiddleware)
	dataRouter.Handle("POST", "/poi", handlers.NewPoiHandler(maxOverpassDistance, overpassUserAgent), sameSiteMiddleware, rateLimitMiddleware)

	rueckenwindServer.AddRouter(rootRouter)
	rueckenwindServer.AddRouter(dataRouter)
	rueckenwindServer.Start()
}
