<script lang="ts">
    import { onDestroy, onMount } from "svelte";

    import ButtonBar from "./components/ButtonBar.svelte";
    import Compass from "./components/Compass.svelte";
    import WeatherInfo from "./components/WeatherInfo.svelte";
    import AboutModal from "./components/AboutModal.svelte";
    import InfoModal from "./components/InfoModal.svelte";
    import PoiDetails from "./components/PoiDetails.svelte";

    import {
        showAboutModal,
        showInfoModal,
        infoTitle,
        infoText,
        userLocation,
        showPoiDetails,
        poisLoading,
        weatherData,
        dataError,
        invalidatePois,
    } from "./stores/store";
    import { fetchWeather, type Location } from "./lib/api";
    import { distance } from "./lib/geo";

    // How far the user has to move before the data fetched for their previous
    // position stops being a good description of where they are. POIs are
    // refetched sooner than the weather, because their bearings and distances
    // are drawn on the compass.
    const poiRefreshDistance = 1; // km
    const weatherRefreshDistance = 5; // km

    // The forecast is published in three hour blocks, but the current block
    // rolls over, so it is refreshed on a timer as well.
    const weatherRefreshInterval = 10 * 60 * 1000; // ms

    let watchId: number | undefined;
    let weatherTimer: number | undefined;

    // Where the data currently on screen was fetched for.
    let weatherFetchedAt: Location | undefined;
    let poisFetchedAt: Location | undefined;

    const loadWeather = async (location: Location) => {
        try {
            $weatherData = await fetchWeather(location);
            weatherFetchedAt = location;
            $dataError = "";
        } catch (error) {
            console.error(error);
            $dataError =
                error instanceof Error
                    ? error.message
                    : "Could not load the forecast.";
        }
    };

    const onLocationUpdate = (position: GeolocationPosition) => {
        const location: Location = {
            lat: position.coords.latitude,
            lon: position.coords.longitude,
        };

        $userLocation = location;

        if (
            weatherFetchedAt === undefined ||
            distance(weatherFetchedAt, location) >= weatherRefreshDistance
        ) {
            void loadWeather(location);
        }

        if (poisFetchedAt === undefined) {
            poisFetchedAt = location;
        } else if (distance(poisFetchedAt, location) >= poiRefreshDistance) {
            poisFetchedAt = location;
            invalidatePois();
        }
    };

    const onLocationFailure = () => {
        $infoTitle = "Permission required";
        $infoText = "This site doesn't work without location permission.";
        $showInfoModal = true;
    };

    onMount(() => {
        if (!navigator.geolocation) {
            $infoTitle = "Location unavailable";
            $infoText = "This browser cannot report your location.";
            $showInfoModal = true;
            return;
        }

        // Watching rather than taking a single fix, because the whole point of
        // the app is that the user is moving.
        watchId = navigator.geolocation.watchPosition(
            onLocationUpdate,
            onLocationFailure,
            {
                enableHighAccuracy: true,
            },
        );

        weatherTimer = window.setInterval(() => {
            if ($userLocation) {
                void loadWeather($userLocation);
            }
        }, weatherRefreshInterval);
    });

    onDestroy(() => {
        if (watchId !== undefined) {
            navigator.geolocation.clearWatch(watchId);
        }

        if (weatherTimer !== undefined) {
            window.clearInterval(weatherTimer);
        }
    });
</script>

<div id="main-container" class="main-container">
    <Compass compassDataWind={$weatherData}/>
    {#if $showPoiDetails && !$poisLoading}
    <PoiDetails></PoiDetails>
    {:else}
    <WeatherInfo weatherData={$weatherData} />
    {/if}
    <ButtonBar />
</div>

{#if $dataError}
    <div class="data-error" role="status">{$dataError}</div>
{/if}

{#if $showAboutModal}
    <AboutModal></AboutModal>
{/if}

{#if $showInfoModal}
    <InfoModal title={$infoTitle} text={$infoText}></InfoModal>
{/if}

<style>
    .data-error {
        position: fixed;
        left: 50%;
        bottom: 1rem;
        transform: translateX(-50%);
        max-width: 90vw;
        padding: 0.5rem 0.75rem;
        border: 1px solid var(--tertiary-warning);
        border-radius: 0.25rem;
        background-color: var(--background);
        color: var(--tertiary-warning);
        font-size: small;
        text-align: center;
        z-index: 2000;
    }
</style>
