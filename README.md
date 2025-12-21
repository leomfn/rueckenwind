# Rückenwind

## Environment variables

- `PORT`: Port number on which the server is running. Default value: 80.
- `STATIC_FILES_DIR`: Path of the static files directory (which contains the index.html and assets directory) relative to the root folder of the application. Default value: './frontend/dist'.
- `OPEN_WEATHER_MAP_API_KEY`
- `DEBUG`: Set to `true` if the program should run in debug mode. This deactivates the tracking middleware.
- `MAX_OVERPASS_DISTANCE`: Maximium distance (in kilometers) to search for POIs. Defaults value: 25.
- `TRACKING_URL`: URL of the Plausible instance's /event endpoint. If not set, deactivate tracking.
- `DOMAIN`: Domain name of the application used for tracking.
