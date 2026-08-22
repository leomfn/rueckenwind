<script lang="ts">
    import { onDestroy, onMount } from "svelte";
    import {
        compassRotation,
        compassStatus,
        pois,
        poiSelectionChoices,
        poisLoading,
        selectedPoi,
        weatherLoading,
    } from "../stores/store";
    import { scale } from "svelte/transition";
    import { backOut } from "svelte/easing";
    import {
        needsPermission,
        requestPermission,
        startCompass,
    } from "../lib/orientation";

    import type { WeatherData } from "../lib/api";

    // Undefined until the first forecast has arrived.
    export let compassDataWind: WeatherData | undefined

    let stopCompass: (() => void) | undefined;

    const listen = () => {
        stopCompass?.();
        stopCompass = startCompass(
            (degrees) => ($compassRotation = degrees),
            (status) => ($compassStatus = status),
        );
    };

    // iOS only delivers readings after the user has allowed it, and only asks
    // when the request comes from a gesture like this one.
    const enableCompass = async () => {
        if (await requestPermission()) {
            $compassStatus = "pending";
            listen();
        } else {
            $compassStatus = "unavailable";
        }
    };

    onMount(() => {
        if (needsPermission()) {
            $compassStatus = "needs-permission";
            return;
        }

        listen();
    });

    onDestroy(() => stopCompass?.());
</script>

{#if $compassStatus === "needs-permission"}
    <button class="compass-permission" on:click={enableCompass}>
        Enable compass
    </button>
{:else if $compassStatus === "unavailable"}
    <div class="compass-permission" role="status">
        Compass unavailable, showing north up
    </div>
{/if}

<div id="compass" class="flex-center" style="rotate: {$compassRotation}deg;">
    <div
        class="compass-circle {$poisLoading || $weatherLoading
            ? 'is-loading'
            : ''} {$compassStatus !== 'active' ? 'not-calibrated' : ''}"
    >
        <div class="direction" id="north">N</div>
        <div class="direction" id="east">E</div>
        <div class="direction" id="south">S</div>
        <div class="direction" id="west">W</div>

        <!-- TODO: fix scaling issues -->
        {#if compassDataWind}
        <div class="arrow future" id="futureWindArrow" style="rotate: {compassDataWind.wind_deg_future}deg; scale: {compassDataWind.wind_scale_future}">
            <svg width="80" height="80" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 23V1 M10 20L12 23L14 20" stroke-width="1" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
        </div>
        <div class="arrow current" id="currentWindArrow" style="rotate: {compassDataWind.wind_deg_current}deg; scale: {compassDataWind.wind_scale_current}">
            <svg width="80" height="80" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 23V1 M10 20L12 23L14 20" stroke-width="1" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
        </div>
        {/if}

        <div
            id="{$selectedPoi}-pois"
            class="sites-container"
            >
            {#each $pois[$selectedPoi] ?? [] as poi, index (poi)}
                <div
                    in:scale={{ duration: 500, easing: backOut }}
                    out:scale={{ duration: 500 }}
                    class="compass-site {$selectedPoi}-poi" style="rotate: {poi.bearing}deg; height: calc(75px + {poi.distance_pixel}px);">
                    <div class="site-text {index === $poiSelectionChoices[$selectedPoi]?.detailsIndex ? 'details-selected' : ''}">{poi.distance_text}</div>
                    <div class="site-indicator {index === $poiSelectionChoices[$selectedPoi]?.detailsIndex ? 'details-selected' : ''}"></div>
                </div>
            {/each}
        </div>
    </div> 
</div>

<style>
    #compass {
        height: 50vh;
        height: 50dvh;
        position: fixed;
        top: 5vh;
        top: 5dvh;
        display: flex;
        justify-content: center;
        align-items: center;
    }

    .compass-circle {
        width: 150px;
        height: 150px;
        border-width: 2px;
        border-style: solid;
        border-color: var(--tertiary-line);
        border-radius: 50%;
        background-color: var(--background);
        box-sizing: border-box;
        position: absolute;
        display: flex;
        justify-content: center;
        align-items: center;
    }

    /* POI markers reach 125px from the centre, so the compass needs 250px of
     * vertical room. A landscape phone gives the compass band about 190px, so
     * everything inside it is scaled down together. Portrait is unaffected. */
    @media (max-height: 500px) {
        .compass-circle {
            transform: scale(0.7);
        }
    }

    /* Marks the compass as not showing a real heading, so that a rose pointing
     * north is not mistaken for a reading. */
    .not-calibrated {
        border-color: var(--tertiary-warning);
    }

    /* The border carries the warning; the text stays in the regular font
     * colour, which is the only one that reads at this size. */
    .compass-permission {
        position: fixed;
        top: 1rem;
        left: 50%;
        transform: translateX(-50%);
        width: auto;
        height: auto;
        min-height: 44px;
        max-width: 90vw;
        padding: 0.4rem 0.75rem;
        border: 1px solid var(--tertiary-warning);
        border-radius: 5px;
        background-color: var(--secondary-background);
        color: var(--font-color);
        font-size: small;
        text-align: center;
        z-index: 1500;
    }

    .direction {
        font-size: 1em;
        height: 1.5em;
        width: 1.5em;
        position: absolute;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    #north {
        left: 50%;
        top: 5%;
        transform: translateX(-50%);
    }

    #east {
        right: 5%;
        top: 50%;
        transform: translateY(-50%);
    }

    #south {
        left: 50%;
        bottom: 5%;
        transform: translateX(-50%);
    }

    #west {
        left: 5%;
        top: 50%;
        transform: translateY(-50%);
    }

    .arrow {
        height: 80px;
        width: 80px;
        position: absolute;
    }

    .sites-container {
        z-index: -10;
    }

    .compass-site {
        color: var(--tertiary);
        /* border-left-color: tomato; */
        position: absolute;
        left: 50%;
        bottom: 50%;
        transform-origin: bottom center;
        font-size: x-small;
        color: var(--font-color);
        z-index: -1;
    }

    .site-text.details-selected {
        top: -0.5rem;
        color: var(--tertiary-warning);
        font-size: medium;
        z-index: 1000;
    }

    .site-indicator.details-selected {
        border-left-width: 2px;
        border-left-color: var(--tertiary-warning);
    }

    .site-text {
        position: fixed;
        top: 0;
        left: 50%;
        transform: translateX(-50%);
    }

    .site-indicator {
        height: calc(100% - 0.8rem);
        width: 0;
        position: fixed;
        bottom: 0;
        left: 50%;
        border-left-width: 1px;
        border-left-style: solid;
        /* border-left-color: var(--tertiary); */
        opacity: 0.3;
    }
</style>