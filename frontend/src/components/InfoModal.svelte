<script lang="ts">
    import { showInfoModal } from "../stores/store";
    import Modal from "./Modal.svelte";

    export let title: string;
    export let text: string;

    const onClose = () => {
        $showInfoModal = false;
    }
</script>

<Modal name="info" onClose={onClose}>
    <h2>{title}</h2>
    <p>
        {text}
    </p>
</Modal>
