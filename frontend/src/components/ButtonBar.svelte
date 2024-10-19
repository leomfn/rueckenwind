<script lang="ts">
    import { showAboutModal, showPoiOptions, userLocation } from "../stores/store";
    import PoiChooser from "./PoiChooser.svelte";

    const openAboutModal = () => {
        $showAboutModal = true;
    }
</script>

<div id="button-bar" class="flex-center">
    <button on:click={openAboutModal}>About</button>

    {#if $userLocation}
    <PoiChooser></PoiChooser>
    {/if}

</div>

<style>
    #button-bar {
        height: 10%;
        position: fixed;
        bottom: 5%;
        left: 50%;
        transform: translateX(-34px);
        display: flex;
        align-items: end;
        gap: 0.5rem;
    }
</style>