<script lang="ts">
    import directionsUrl from "../../static/images/route.svg";
    import websiteUrl from "../../static/images/globe.svg";

    import {
        pois,
        poiSelectionChoices,
        selectedPoi,
        userLocation,
    } from "../stores/store";
    import type { poiElement } from "../types/types";

    let detailWebsite: string = "";
    let detailDirections: string = "";

    const selectPoiDetails = (poi: poiElement, index: number) => {
        if (index === $poiSelectionChoices[$selectedPoi].detailsIndex) {
            $poiSelectionChoices[$selectedPoi].detailsIndex = undefined;
            detailWebsite = "";
            detailDirections = "";
        } else {
            $poiSelectionChoices[$selectedPoi].detailsIndex = index;
            detailWebsite = poi.website;
            detailDirections = `https://brouter.de/brouter-web/#map=14/${($userLocation.lat + poi.lat) / 2}/${($userLocation.lon + poi.lon) / 2}/standard&lonlats=${$userLocation.lon},${$userLocation.lat};${poi.lon},${poi.lat}&profile=safety`;
        }
    };
</script>

<div class="poi-details">
    <div class="poi-details-table-container">
        <table>
            <tbody>
                {#each $pois[$selectedPoi] as poi, index (poi)}
                    <tr
                        class="poi-details-item {index ===
                        $poiSelectionChoices[$selectedPoi].detailsIndex
                            ? 'details-selected'
                            : ''}"
                        on:click={() => selectPoiDetails(poi, index)}
                    >
                        <td>{poi.distance_text} km</td>
                        <td>{poi.name ? poi.name : ""}</td>
                        <td>{poi.address ? poi.address : ""}</td>
                    </tr>
                {/each}
            </tbody>
        </table>
    </div>
    {#if $poiSelectionChoices[$selectedPoi].detailsIndex != undefined}
        <div class="detail-links">
            <a
                class="button"
                href={detailDirections}
                rel="noreferrer noopener"
                target="_blank"
                data-umami-event="poi-get-directions"
            >
                <img src={directionsUrl} alt="" />
                <span>Directions</span>
            </a>
            {#if detailWebsite != ""}
                <a
                    class="button"
                    href={detailWebsite}
                    rel="noreferrer noopener"
                    target="_blank"
                    data-umami-event="poi-visit-website"
                >
                    <img src={websiteUrl} alt="" />
                    <span>Website</span>
                </a>
            {/if}
        </div>
    {/if}
</div>

<style>
    .poi-details {
        height: 30%;
        width: 80%;
        position: fixed;
        bottom: 15%;
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }

    .detail-links {
        display: flex;
        gap: 0.5rem;
    }

    table {
        border-collapse: collapse;
        width: 100%;
        font-size: smaller;
    }

    td {
        padding: 0.1rem 0.5rem;
    }

    a.button {
        height: 1rem;
        width: auto;
        display: flex;
        align-items: center;
        gap: 0.2rem;
        font-size: x-small;
        padding: 0.2rem 0.24rem;
        font-family: "Open Sans", sans-serif;

        border-style: solid;
        border-width: 1px;
        border-color: var(--tertiary);
        border-radius: 8px;
        background-color: var(--background);
        color: var(--font-color);

        text-decoration: none;
        cursor: pointer;
    }

    a.button:active {
        background-color: var(--tertiary);
    }

    img {
        height: 0.8rem;
        width: 0.8rem;
    }

    .poi-details-table-container {
        max-height: 70%;
        font-size: 1rem;
        border: solid 1px black;
        border-radius: 5px;
        box-sizing: border-box;
        overflow: auto;
    }

    .poi-details-item {
        white-space: nowrap;
        width: 100%;
        user-select: none;
    }

    .poi-details-item:hover,
    .details-selected {
        background-color: var(--tertiary-warning);
    }
</style>
