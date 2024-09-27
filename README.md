# Windy

## Environment variables

- **OPEN_WEATHER_MAP_API_KEY**
- **JWT_SECRET**: Secret used to sign JWT tokens. Create e. g. using `openssl rand -hex 32`.
- **PROXY**: Set to `true` if application is running behind a proxy. In this case, the X-Forward-For header will be use for IP address logging. Note that this might be vulnerable to IP spoofing.
