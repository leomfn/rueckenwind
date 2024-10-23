<script lang="ts">
    import detailsUrl from '../../static/images/directions.svg';
    import xUrl from '../../static/images/x.svg';

    import { poisLoading, selectedPoi, showAboutModal, showPoiDetails, userLocation } from "../stores/store";
    import PoiChooser from "./PoiChooser.svelte";

    const openAboutModal = () => {
        $showAboutModal = true;
    }

    const togglePoiDetails = () => {
        $showPoiDetails  =!$showPoiDetails;
    }
</script>

<div id="button-bar" class="flex-center">
    {#if $selectedPoi && !$poisLoading}
    <button class="details-button" on:click={togglePoiDetails}>
        <img src={$showPoiDetails ? xUrl : detailsUrl} alt="">
    </button>
    {/if}
    <button class="about-button" on:click={openAboutModal}>About</button>
    {#if $userLocation}
    <PoiChooser></PoiChooser>
    {/if}

</div>

<style>
    #button-bar {
        height: 10%;
        position: fixed;
        bottom: 5%;
        display: flex;
        align-items: end;
        justify-content: center;
    }

    .details-button {
        position: absolute;
        transform: translateX(calc(-34px - 20px - 0.5rem));
        width: 40px;
        height: 40px;
        background-color: var(--background);
    }

    .details-button:active {
        background-color: var(--tertiary);
        border-color: var(--tertiary);
    }

    .details-button img {
        position: absolute;
        left: 50%;
        top: 50%;
        transform: translate(-50%, -50%);
        width: 50%;
        height: 50%;
    }
</style>