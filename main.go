package main

func main() {
	server := newServer()

	rootRouter := newRouter("/")
	rootRouter.handle("GET", "/{$}", newGetIndexHandler(), loggingMiddleware, trackingMiddleware)
	rootRouter.handle("GET", "/static/", newStaticFilesHandler(staticFilesDir))
	rootRouter.handle("GET", "/about", newAboutHandler(), loggingMiddleware, sameSiteMiddleware, trackingMiddleware)
	rootRouter.handle("GET", "/error", newErrorHandler(), loggingMiddleware, sameSiteMiddleware, trackingMiddleware)

	dataRouter := newRouter("localhost/data/")
	dataRouter.handle("POST", "/weather", newWeatherHandler(), loggingMiddleware, sameSiteMiddleware, trackingMiddleware)
	dataRouter.handle("POST", "/sites", newSitesHandler(), loggingMiddleware, sameSiteMiddleware, trackingMiddleware)

	server.addRouter(rootRouter)
	server.addRouter(dataRouter)
	server.start()
}
