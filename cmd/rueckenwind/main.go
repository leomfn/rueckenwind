package main

import (
	"github.com/leomfn/rueckenwind/internal/handlers"
	"github.com/leomfn/rueckenwind/internal/middleware"
	"github.com/leomfn/rueckenwind/internal/server"
)

func main() {
	rueckenwindServer := server.NewServer(port)

	trackingMiddleware := middleware.NewTrackingMiddleware(domain, trackingUrl, debug)
	sameSiteMiddleware := middleware.NewSameSiteMiddleware(domain, debug)

	rootRouter := server.NewRouter("/")
	rootRouter.Handle("GET", "/{$}", handlers.NewGetIndexHandler(), trackingMiddleware)
	rootRouter.Handle("GET", "/static/", handlers.NewStaticFilesHandler(staticFilesDir))
	rootRouter.Handle("GET", "/about", handlers.NewAboutHandler(), sameSiteMiddleware, trackingMiddleware)
	rootRouter.Handle("GET", "/error", handlers.NewErrorHandler(), sameSiteMiddleware, trackingMiddleware)

	dataRouter := server.NewRouter("/data/")
	dataRouter.Handle("POST", "/weather", handlers.NewWeatherHandler(owmApiKey), sameSiteMiddleware, trackingMiddleware)
	dataRouter.Handle("POST", "/poi", handlers.NewPoiHandler(maxOverpassDistance), sameSiteMiddleware, trackingMiddleware)

	rueckenwindServer.AddRouter(rootRouter)
	rueckenwindServer.AddRouter(dataRouter)
	rueckenwindServer.Start()
}
