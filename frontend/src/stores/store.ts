import { writable } from "svelte/store";
import type { Pois } from "../types/types";

import campsiteUrl from '../../static/images/campsite.svg';
import waterUrl from '../../static/images/water.svg';
import coffeeUrl from '../../static/images/coffee.svg';
import observationUrl from '../../static/images/observation.svg';

export const showAboutModal = writable<boolean>(false);
export const showInfoModal = writable<boolean>(false);

export const showPoiDetails = writable<boolean>(false);
// export const selectedPoiDetailsIndex = writable<number>();

export const infoTitle = writable<string>("Example Title");
export const infoText = writable<string>("Example Description");

export const showPoiOptions = writable<boolean>(false);
export const selectedPoi = writable<string>();
export const previouslySelectedPoi = writable<string>();

export const pois = writable<Pois>({});

export const poisLoading = writable<boolean>(false);

export const userLocation = writable<{lat: number, lon: number}>();

export const poiSelectionChoices = writable<Record<string, {img: string, detailsIndex?: number}>>({
    camping: {
        img: campsiteUrl,
    },
    water: {
        img: waterUrl
    },
    cafe: {
        img: coffeeUrl
    },
    observation: {
        img: observationUrl
    }
})
