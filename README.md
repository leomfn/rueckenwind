# Rückenwind

## Environment variables

- `OPEN_WEATHER_MAP_API_KEY`
- `PROXY`: Set to `true` if application is running behind a proxy. In this case, the X-Forward-For header will be use for IP address logging. Note that this might be vulnerable to IP spoofing.
- `DEBUG`: Set to `true` if the program should run in debug mode. This deactivates the tracking middleware.
- `MAX_OVERPASS_DISTANCE`: Maximium distance (in kilometers) to search for POIs. Defaults to 25km.
- `MAX_SITES`: Maximum number of sites to return for each POI type. Defaults to 10.
- `TRACKING_URL`: URL of the Plausible instance's /event endpoint.
- `DOMAIN`: Domain name of the application used for tracking.
