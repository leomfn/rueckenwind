<script lang="ts">
    export let name: String;
    export let onClose;

</script>

<div id="{name}-modal" class="modal">
    <div class="modal-content">
        <slot></slot>
    </div>
    <div on:click={onClose} class="modal-underlay"></div>
    <div class="modal-close-container">
        <button id="button-close-{name}-modal" class="button-close-modal close-button" on:click={onClose}>
            Close
        </button>
    </div>
</div>

<style>
    .modal {
        height: 100%;
        width: 100%;
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        align-items: center;
        position: fixed;
        background-color: rgba(0, 0, 0, 0.3);
        backdrop-filter: blur(5px);
    }

    .modal-underlay {
        position: absolute;
        top: 0;
        bottom: 0;
        left: 0;
        right: 0;
        z-index: 0;
    }

    .modal-content {
        width: 80%;
        max-width: 600px;
        max-height: 75%;
        margin-top: 10vh;
        overflow-y: scroll;
        border: solid 1px var(--primary);
        border-radius: 5px;
        box-shadow: 0px 0px 20px 0px rgba(0, 0, 0, 0.3);
        background-color: white;
        padding: 0 20px 0px 20px;
        user-select: text;
        -webkit-user-select: text;
        z-index: 1;
    }

    .modal-content :global(h2) {
        margin: 0;
        padding-top: 1rem;
        padding-bottom: 0.5rem;
        position: sticky;
        top: 0;
        background-color: white;
    }

    .modal-close-container {
        height: 10%;
        position: fixed;
        bottom: 5%;
        display: flex;
        justify-content: center;
        align-items: end;
        z-index: 1;
    }
</style>
