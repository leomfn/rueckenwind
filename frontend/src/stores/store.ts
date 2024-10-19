import { writable } from "svelte/store";

export const showAboutModal = writable<boolean>(false);
export const showInfoModal = writable<boolean>(false);

export const infoTitle = writable<string>("Example Title");
export const infoText = writable<string>("Example Description");

export const showPoiOptions = writable<boolean>(false);
export const selectedPoi = writable<string>();

export const pois = writable<Record<string, any>>({});

export const userLocation = writable<{lat: number, lon: number}>();
