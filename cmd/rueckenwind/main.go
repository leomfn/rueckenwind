package main

import (
	"fmt"

	"github.com/leomfn/rueckenwind/internal/handlers"
	"github.com/leomfn/rueckenwind/internal/middleware"
	"github.com/leomfn/rueckenwind/internal/server"
)

func main() {
	rueckenwindServer := server.NewServer(port)

	trackingMiddleware := middleware.NewTrackingMiddleware(domain, trackingUrl, debug)
	sameSiteMiddleware := middleware.NewSameSiteMiddleware(domain, debug)

	rootRouter := server.NewRouter("/")
	rootRouter.Handle("GET", "/{$}", handlers.NewGetIndexHandler(fmt.Sprintf("%s/index.html", staticFilesDir)), trackingMiddleware)
	rootRouter.Handle("GET", "/assets/", handlers.NewStaticFilesHandler(fmt.Sprintf("%s/assets", staticFilesDir)), sameSiteMiddleware)
	rootRouter.Handle("GET", "/health", handlers.NewHealthcheckHandler())

	dataRouter := server.NewRouter("/data/")
	dataRouter.Handle("POST", "/weather", handlers.NewWeatherHandler(owmApiKey), sameSiteMiddleware, trackingMiddleware)
	dataRouter.Handle("POST", "/poi", handlers.NewPoiHandler(maxOverpassDistance), sameSiteMiddleware, trackingMiddleware)

	rueckenwindServer.AddRouter(rootRouter)
	rueckenwindServer.AddRouter(dataRouter)
	rueckenwindServer.Start()
}
