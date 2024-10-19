<script lang="ts">
    import { showAboutModal } from "../stores/store";
    import Modal from "./Modal.svelte";

    const onClose = () => {
        $showAboutModal = false;
    }
</script>

<Modal name="about" onClose={onClose}>
    <h2>About Rückenwind</h2>
    <p>
        This site shows you <span class="current"><b>current</b></span> and <span class="future"><b>future</b></span> (3 hours from now) weather information for your location.
        It focuses on wind direction and speed, so you can make decisions during your bike adventure.
        The small numbers around the compass represent distances (in km) to the nearest Points of Interest (POI).
        The current POI category can be changed using the button on the bottom right.
    </p>

    <h3>Weather legend</h3>
    <ul id="symbol-legend">
        <li class="legend-item"><img src="/static/images/thermometer.svg">temperature</li>
        <li class="legend-item"><img src="/static/images/wind.svg">wind speed</li>
        <li class="legend-item"><img src="/static/images/tornado.svg">gust speed</li>
        <li class="legend-item"><img src="/static/images/cloud-rain-wind.svg">rain intensity</li>
        <li class="legend-item"><img src="/static/images/sunset.svg">sunset time</li>
    </ul>

    <h3>POI legend</h3>
    <ul id="symbol-legend">
        <li class="legend-item"><img src="/static/images/campsite.svg">campsites</li>
        <li class="legend-item"><img src="/static/images/water.svg">drinking water</li>
        <li class="legend-item"><img src="/static/images/coffee.svg">cafes</li>
        <li class="legend-item"><img src="/static/images/observation.svg">observation points</li>
    </ul>

    <h2>Permissions</h2>
    <p>
        This site requires access to your device's <em>location</em> to fetch weather data and POIs around you.
    </p>
    <p>
        If you're using a mobile device, your device's <em>orientation</em> sensor is used to rotate the compass.
        On iOS, you have to give this permission explicitly.
    </p>

    <h2>Privacy</h2>
    <p>
        The coordinates from your device's location are shared with OpenWeather and Overpass API (see below).
        However, no further data from your device is sent to those services.
    </p>
    <p>
        To track usage statistics, this site uses a self-hosted <a href="https://github.com/plausible/community-edition" rel="noreferrer noopener" target="_blank">Plausible</a> instance.
        This way, information about your browser type and version is stored, along with your anonymized IP address and the date and time you accessed this site.
    </p>
    <p>
        This site doesn't use cookies.
    </p>

    <h2>Data sources</h2>
    <p>
        Weather data is provided by <a href="https://openweathermap.org/" rel="noreferrer noopener" target="_blank">OpenWeather</a>.
        POI locations are fetched from <a href="https://www.openstreetmap.org" rel="noreferrer noopener" target="_blank">OpenStreetMap</a> via <a href="https://wiki.openstreetmap.org/wiki/Overpass_API" rel="noreferrer noopener" target="_blank">Overpass API</a>.
    </p>
    <div class="source-logos">
        <a href="https://openweathermap.org" rel="noreferrer noopener" target="_blank">
            <img src="/static/images/open_weather_logo.svg" alt="OpenWeather logo" height="40px">
        </a>
    </div>

    <h2>Icons</h2>
    <p>
        The icons on this site are provided by <a href="https://lucide.dev" rel="noreferrer noopener" target="_blank">Lucide</a> and are licensed under ISC.
    </p>

    <h2>Source code</h2>
    <p>
        You can find the source on <a href="https://github.com/leomfn/rueckenwind" rel="noreferrer noopener" target="_blank">GitHub</a>.
        Feel free to open an issue if you find a bug.
    </p>

    <h2>Disclaimer</h2>
    <p>
        Please note that the information provided on this website is for general informational purposes only and may not be fully accurate or up-to-date.
        You are responsible for how you interpret and use this information.
        Always verify details independently.
    </p>

</Modal>

<style>
    #symbol-legend {
        list-style-type: none;
    }

    .legend-item {
        display: flex;
        align-items: center;
        margin-bottom: 0.5rem;
    }

    .legend-item img {
        padding-right: 1rem;
    }
</style>
