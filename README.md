# Rückenwind

## Environment variables

- `PORT`: Port number on which the server is running.
- `STATIC_FILES_DIR`: Path of the static files directory relative to the root folder of the application.
- `OPEN_WEATHER_MAP_API_KEY`
- `PROXY`: Set to `true` if application is running behind a proxy. In this case, the X-Forward-For header will be use for IP address logging. Note that this might be vulnerable to IP spoofing.
- `DEBUG`: Set to `true` if the program should run in debug mode. This deactivates the tracking middleware.
- `MAX_OVERPASS_DISTANCE`: Maximium distance (in kilometers) to search for POIs. Defaults to 25km.
- `TRACKING_URL`: URL of the Plausible instance's /event endpoint.
- `DOMAIN`: Domain name of the application used for tracking.
