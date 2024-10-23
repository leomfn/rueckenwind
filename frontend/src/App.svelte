<script lang="ts">
    import { onMount } from "svelte";

    import ButtonBar from "./components/ButtonBar.svelte";
    import Compass from "./components/Compass.svelte";
    import WeatherInfo from "./components/WeatherInfo.svelte";
    import AboutModal from "./components/AboutModal.svelte";
    import InfoModal from "./components/InfoModal.svelte";

    import { showAboutModal, showInfoModal, infoTitle, infoText, userLocation, showPoiDetails, poisLoading } from "./stores/store";
    import PoiDetails from "./components/PoiDetails.svelte";

    // load weather data
    interface WeatherData {
        temp_current: number;
        wind_current: number;
        wind_gust_current: number;
        wind_deg_current: number;
        wind_scale_current: number;
        rain_current_text: string;

        temp_future: number;
        wind_future: number;
        wind_gust_future: number;
        wind_deg_future: number;
        wind_scale_future: number;
        rain_future_text: string;

        sunset: string;
    }

    let weatherData: WeatherData;

    const getData = () => {
        fetch("/data/weather", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                lon: $userLocation.lon,
                lat: $userLocation.lat,
            }),
        })
            .then((res) => res.json())
            .then((data) => {
                weatherData = data;
            });
    };

    const getUserLocation = () => {
        const locationSuccess = (position: GeolocationPosition) => {
            userLocation.set({
                lat: position.coords.latitude,
                lon: position.coords.longitude
            })

            getData();
        }

        const locationFailure = () => {
            $infoTitle = "Permission required";
            $infoText = "This site doesn't work without location permission.";
            $showInfoModal = true;
        }

        if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition(
                locationSuccess,
                locationFailure,
                {
                    enableHighAccuracy: true,
                }
            )
        }
    }

    onMount(() => {
        getUserLocation();
    });

</script>

<div id="main-container" class="main-container">
    <Compass compassDataWind={weatherData}/>
    {#if $showPoiDetails && !$poisLoading}
    <PoiDetails></PoiDetails>
    {:else}
    <WeatherInfo weatherData={weatherData} />
    {/if}
    <ButtonBar />
</div>

{#if $showAboutModal}
    <AboutModal></AboutModal>
{/if}

{#if $showInfoModal}
    <InfoModal title={$infoTitle} text={$infoText}></InfoModal>
{/if}
