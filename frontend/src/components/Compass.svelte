<script lang="ts">
    import { onMount } from "svelte";
    import { pois, poisLoading, selectedPoi } from "../stores/store";
    import { scale } from "svelte/transition";
    import { backOut } from "svelte/easing";

    interface CompassDataWind {
        wind_deg_current: number;
        wind_scale_current: number;

        wind_deg_future: number;
        wind_scale_future: number;
    }

    export let compassDataWind: CompassDataWind

    let orientationDegrees: number = 0;

    onMount(() => {
        // TODO: differentiate iOS and Android
        // TODO: add modal for iOS to give permission
        window.addEventListener('deviceorientationabsolute', event => {
            if (event.alpha != null) {
                orientationDegrees = event.alpha;
            } else {
                orientationDegrees = 0;
            }
        })
    })
</script>

<div id="compass" class="flex-center" style="rotate: {orientationDegrees}deg;">
    <div class="compass-circle {$poisLoading ? 'sites-loading' : ''}">
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
            {#each $pois[$selectedPoi] as poi (poi)}
                <div
                    in:scale={{ duration: 500, easing: backOut }}
                    out:scale={{ duration: 500 }}
                    class="compass-site {$selectedPoi}-poi" style="rotate: {poi.bearing}deg; height: calc(75px + {poi.distance_pixel}px);">
                    <div class="site-text">{poi.distance_text}</div>
                    <div class="site-indicator"></div>
                </div>
            {/each}
        </div>
    </div> 
</div>

<style>
    #compass {
        height: 50%;
        position: fixed;
        top: 5%;
        display: flex;
        justify-content: center;
        align-items: center;
    }

    .compass-circle {
        width: 150px;
        height: 150px;
        border-width: 2px;
        border-style: solid;
        border-color: var(--tertiary);
        border-radius: 50%;
        background-color: var(--background);
        box-sizing: border-box;
        position: absolute;
        display: flex;
        justify-content: center;
        align-items: center;
    }

    /* TODO: Currently, there is no way to detect magnetometer calibration. Maybe
     * find an alternative. */
    .not-calibrated {
        border-color: var(--tertiary-warning);
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
        position: absolute;
        left: 50%;
        bottom: 50%;
        transform-origin: bottom center;
        font-size: x-small;
        color: var(--font-color);
        text-align: center;
        z-index: -1;
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
        border-left: 1px dotted black;
    }
</style>