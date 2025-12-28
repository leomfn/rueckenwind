<script lang="ts">
    import searchUrl from "../../static/images/search.svg";
    import xUrl from "../../static/images/x.svg";

    import {
        pois,
        poiSelectionChoices,
        poisLoading,
        previouslySelectedPoi,
        selectedPoi,
        showPoiOptions,
        userLocation,
    } from "../stores/store";

    const togglePoiOptions = () => {
        showPoiOptions.update((value) => !value);
    };

    const selectPoi = (poi: string) => {
        $previouslySelectedPoi = $selectedPoi;
        $selectedPoi = poi;
        $showPoiOptions = false;

        $poisLoading = true;

        // Check if pois have been fetched before
        if (poi in $pois) {
            $poisLoading = false;
            return;
        }

        // Track if umami has loaded successfully
        window.umami?.track(`poi-${poi}`);

        fetch("/data/poi", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                category: poi,
                lon: $userLocation.lon,
                lat: $userLocation.lat,
            }),
        })
            .then((res) => res.json())
            .then((data) => {
                pois.update((current) => {
                    return { ...current, [poi]: data };
                });
            })
            .then(() => {
                $poisLoading = false;
            });
    };

    const chooserSymbol = (showPoiOptions: boolean): string => {
        if (showPoiOptions) {
            return xUrl;
        } else if ($selectedPoi in $poiSelectionChoices) {
            return $poiSelectionChoices[$selectedPoi].img;
        }

        return searchUrl;
    };

    let chooserImageSrc: string = chooserSymbol($showPoiOptions);
    $: chooserImageSrc = chooserSymbol($showPoiOptions);
</script>

<div id="sites-fab-container">
    <div>
        <button
            id="sites-fab-main"
            class="sites-fab {$poisLoading ? 'sites-loading' : ''}"
            on:click={togglePoiOptions}
        >
            <img
                id="sites-fab-main-image"
                src={chooserImageSrc}
                alt="Point of interest chooser"
            />
        </button>
    </div>
    {#if $showPoiOptions}
        {#each Object.entries($poiSelectionChoices) as [title, config]}
            <div>
                <button
                    id="sites-fab-{title}"
                    class="sites-fab sites-fab-choices {title === $selectedPoi
                        ? 'sites-fab-selected'
                        : ''}"
                    on:click={() => selectPoi(title)}
                >
                    <img src={config.img} alt={title} />
                </button>
            </div>
        {/each}
    {/if}
</div>

<style>
    #sites-fab-container {
        position: absolute;
        transform: translateX(calc(34px + 20px + 0.5rem));
        display: flex;
        flex-direction: column-reverse;
        align-items: start;
        gap: 0.3rem;
    }

    .sites-fab {
        position: relative;
        width: 40px;
        height: 40px;
        background-color: var(--background);
    }

    .sites-fab:active {
        background-color: var(--tertiary);
        border-color: var(--tertiary);
    }

    .sites-fab-choices {
        opacity: 1;
        transition: opacity 0.1s ease;
    }

    .sites-fab-selected {
        background-color: var(--tertiary);
    }

    .sites-fab-selected:active {
        background-color: var(--background);
    }

    .sites-fab img {
        position: absolute;
        left: 50%;
        top: 50%;
        transform: translate(-50%, -50%);
        width: 50%;
        height: 50%;
    }
</style>
