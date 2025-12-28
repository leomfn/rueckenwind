<script lang="ts">
    import thermometerUrl from '../../static/images/temperature.svg';
    import windUrl from '../../static/images/wind.svg';
    import gustUrl from '../../static/images/gust.svg';
    import rainUrl from '../../static/images/rain.svg';
    import sunsetUrl from '../../static/images/sunset.svg';

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
        <img src={thermometerUrl} alt="Current and future temperatures in degrees celsius"><span class="current">{weatherData.temp_current}</span> <span class="future">{weatherData.temp_future}</span> °C
    </div>
    <div class="info-row">
        <img src={windUrl} alt="Current and future wind speed in kilometers per hour"><span class="current">{weatherData.wind_current}</span> <span class="future">{weatherData.wind_future}</span> km/h
    </div>
    <div class="info-row">
        <img src={gustUrl} alt="Current and future gust speed in kilometers per hour"><span class="current">{weatherData.wind_gust_current}</span> <span class="future">{weatherData.wind_gust_future}</span> km/h
    </div>
    <div class="info-row">
        <img src={rainUrl} alt="Current and future precipitation"><span class="current">{weatherData.rain_current_text}</span> <span class="future">{weatherData.rain_future_text}</span>
    </div>
    <div class="info-row info-sunset">
        <img src={sunsetUrl} alt="Today's sunset time"><span class="current">{weatherData.sunset}</span>
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

    .info-sunset {
        margin-top: 1rem;
    }
</style>
