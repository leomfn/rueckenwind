# Rückenwind

## Environment variables

- `PORT`: Port number on which the server is running. Default value: 80.
- `STATIC_FILES_DIR`: Path of the static files directory (which contains the index.html and assets directory) relative to the root folder of the application. Default value: './frontend/dist'.
- `OPEN_WEATHER_MAP_API_KEY`
- `DEBUG`: Set to `true` if the program should run in debug mode. This deactivates the tracking middleware.
- `MAX_OVERPASS_DISTANCE`: Maximium distance (in kilometers) to search for POIs. Defaults value: 25.
- `OVERPASS_USER_AGENT`: Custom user agent sent with requests to the Overpass API. Its usage rules require a custom, identifying user agent (ideally with contact info); stock or faked user agents may get blocked. Default value: 'rueckenwind'.
- `DOMAIN`: Domain name of the application.
- `VITE_TRACKING_URL`: URL of the Umami instance.
- `VITE_TRACKING_ID`: Website-ID of the Umami website configuration.
