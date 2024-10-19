<script lang="ts">
    interface WeatherData {
        temp_current: number;
        wind_current: number;
        wind_gust_current: number;
        rain_current_text: string;

        temp_future: number;
        wind_future: number;
        wind_gust_future: number;
        rain_future_text: string;

        sunset: string;
    }

    export let weatherData: WeatherData;
</script>

{#if weatherData}
<div id="weather-info" class="grid-item">
    <div class="info-row">
        <img src="/static/images/thermometer.svg"><span class="current">{weatherData.temp_current}</span> <span class="future">{weatherData.temp_future}</span> °C
    </div>
    <div class="info-row">
        <img src="/static/images/wind.svg"><span class="current">{weatherData.wind_current}</span> <span class="future">{weatherData.wind_future}</span> km/h
    </div>
    <div class="info-row">
        <img src="/static/images/tornado.svg"><span class="current">{weatherData.wind_gust_current}</span> <span class="future">{weatherData.wind_gust_future}</span> km/h
    </div>
    <div class="info-row">
        <img src="/static/images/cloud-rain-wind.svg"><span class="current">{weatherData.rain_current_text}</span> <span class="future">{weatherData.rain_future_text}</span>
    </div>
    <div class="info-row">
        <img src="/static/images/sunset.svg"><span class="current">{weatherData.sunset}</span>
    </div>
</div>
{/if}

<style>
    #weather-info {
        height: 30%;
        position: fixed;
        bottom: 15%;
        font-size: 1rem;
        display: inline-flex;
        flex-direction: column;
        justify-content: center;
        align-items: start;
    }

    .info-row {
        display: flex;
        align-items: center;
        margin-bottom: 0.2rem;
    }

    .info-row img {
        padding-right: 1rem;
        height: 1rem;
    }

    .info-row .current {
        margin-right: 0.5rem;
    }

    .info-row .future {
        margin-right: 0.5rem;
    }
</style>