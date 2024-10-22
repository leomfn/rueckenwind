import { writable } from "svelte/store";
import type { Pois } from "../types/types";

export const showAboutModal = writable<boolean>(false);
export const showInfoModal = writable<boolean>(false);

export const infoTitle = writable<string>("Example Title");
export const infoText = writable<string>("Example Description");

export const showPoiOptions = writable<boolean>(false);
export const selectedPoi = writable<string>();
export const previouslySelectedPoi = writable<string>();

export const pois = writable<Pois>({});

export const poisLoading = writable<boolean>(false);

export const userLocation = writable<{lat: number, lon: number}>();
