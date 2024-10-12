package main

func main() {
	server := newServer()

	rootRouter := newRouter("/")
	rootRouter.handle("GET", "/{$}", newGetIndexHandler(), trackingMiddleware)
	rootRouter.handle("GET", "/static/", newStaticFilesHandler(staticFilesDir))
	rootRouter.handle("GET", "/about", newAboutHandler(), sameSiteMiddleware, trackingMiddleware)
	rootRouter.handle("GET", "/error", newErrorHandler(), sameSiteMiddleware, trackingMiddleware)

	dataRouter := newRouter("localhost/data/")
	dataRouter.handle("POST", "/weather", newWeatherHandler(), sameSiteMiddleware, trackingMiddleware)
	dataRouter.handle("POST", "/sites", newSitesHandler(), sameSiteMiddleware, trackingMiddleware)

	server.addRouter(rootRouter)
	server.addRouter(dataRouter)
	server.start()
}
