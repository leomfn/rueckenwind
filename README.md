# Rückenwind

## Environment variables

- `OPEN_WEATHER_MAP_API_KEY`
- `PROXY`: Set to `true` if application is running behind a proxy. In this case, the X-Forward-For header will be use for IP address logging. Note that this might be vulnerable to IP spoofing.
- `MAX_OVERPASS_DISTANCE`: Maximium distance (in kilometers) to search for POIs. Defaults to 25km.
