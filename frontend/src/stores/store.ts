import { get, writable } from "svelte/store";
import type { Pois } from "../types/types";
import type { CompassStatus } from "../lib/orientation";
import type { Location, WeatherData } from "../lib/api";
import { fetchPois } from "../lib/api";

import campsiteUrl from '../../static/images/campsite.svg';
import waterUrl from '../../static/images/water.svg';
import coffeeUrl from '../../static/images/coffee.svg';
import observationUrl from '../../static/images/observation.svg';

export const showAboutModal = writable<boolean>(false);
export const showInfoModal = writable<boolean>(false);

export const showPoiDetails = writable<boolean>(false);

export const infoTitle = writable<string>("Example Title");
export const infoText = writable<string>("Example Description");

export const showPoiOptions = writable<boolean>(false);
export const selectedPoi = writable<string>();
export const previouslySelectedPoi = writable<string>();

export const pois = writable<Pois>({});

export const poisLoading = writable<boolean>(false);

export const userLocation = writable<Location>();

export const weatherData = writable<WeatherData | undefined>();

// Set when a request fails, so that a failure is visible instead of leaving the
// display silently empty.
export const dataError = writable<string>("");

// How long a message stays on screen before it removes itself.
const errorDismissDelay = 8000;

let errorTimer: number | undefined;

// Shows a message and schedules its removal. A failure is often not followed by
// a success, so the message has to expire on its own rather than waiting for
// one to clear it.
export const showError = (message: string): void => {
    if (errorTimer !== undefined) {
        window.clearTimeout(errorTimer);
    }

    dataError.set(message);

    errorTimer = window.setTimeout(() => {
        errorTimer = undefined;
        dataError.set("");
    }, errorDismissDelay);
};

export const clearError = (): void => {
    if (errorTimer !== undefined) {
        window.clearTimeout(errorTimer);
        errorTimer = undefined;
    }

    dataError.set("");
};

export const compassStatus = writable<CompassStatus>("pending");
export const compassRotation = writable<number>(0);

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

// Loads the POIs of a category, reusing what has already been fetched for the
// current location.
export const loadPois = async (category: string): Promise<void> => {
    const location = get(userLocation);

    if (!location) {
        return;
    }

    if (category in get(pois)) {
        return;
    }

    poisLoading.set(true);

    // This is a deliberate retry when a previous attempt failed, so any message
    // still on screen is stale.
    clearError();

    try {
        const found = await fetchPois(category, location);
        pois.update((current) => ({ ...current, [category]: found }));
    } catch (error) {
        console.error(error);
        showError(
            error instanceof Error ? error.message : "Could not load places.",
        );
    } finally {
        poisLoading.set(false);
    }
};

// Drops POIs fetched for a previous location. Distances and bearings are
// relative to where the user was when they were fetched, so keeping them after
// the user has moved would point the compass at the wrong places.
export const invalidatePois = (): void => {
    pois.set({});

    poiSelectionChoices.update((choices) => {
        for (const choice of Object.values(choices)) {
            choice.detailsIndex = undefined;
        }

        return choices;
    });

    const category = get(selectedPoi);
    if (category) {
        void loadPois(category);
    }
};
