# Rückenwind

## Environment variables

- `PORT`: Port number on which the server is running. Default value: 80.
- `STATIC_FILES_DIR`: Path of the static files directory (which contains the index.html and assets directory) relative to the root folder of the application. Default value: './frontend/dist'.
- `OPEN_WEATHER_MAP_API_KEY`
- `DEBUG`: Set to `true` if the program should run in debug mode. This deactivates the tracking middleware.
- `MAX_OVERPASS_DISTANCE`: Maximium distance (in kilometers) to search for POIs. Defaults value: 25.
- `OVERPASS_USER_AGENT`: Custom user agent sent with requests to the Overpass API. Its usage rules require a custom, identifying user agent (ideally with contact info); stock or faked user agents may get blocked. Default value: 'rueckenwind'.
- `RATE_LIMIT_PER_MINUTE`: Requests per minute allowed per client on the data endpoints. Default value: 60.
- `RATE_LIMIT_BURST`: Number of requests a client may make back-to-back before the rate limit applies. Default value: 15.
- `TRUST_PROXY_HEADERS`: Set to `true` to identify clients by the `X-Forwarded-For` header instead of the connection's remote address. Only enable this when the application sits behind a reverse proxy that sets the header, otherwise clients can spoof it to bypass the rate limit. Default value: `false`.
- `DOMAIN`: Domain name of the application.
- `VITE_TRACKING_URL`: URL of the Umami instance.
- `VITE_TRACKING_ID`: Website-ID of the Umami website configuration.
