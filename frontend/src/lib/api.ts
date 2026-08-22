import type { poiElement } from "../types/types";

export interface Location {
    lat: number;
    lon: number;
}

export interface WeatherData {
    temp_current: number;
    wind_current: number;
    wind_gust_current: number;
    wind_deg_current: number;
    wind_scale_current: number;
    rain_current_text: string;

    temp_future: number;
    wind_future: number;
    wind_gust_future: number;
    wind_deg_future: number;
    wind_scale_future: number;
    rain_future_text: string;

    sunset: string;
}

// Posts a JSON body and returns the decoded response. A non-OK status is turned
// into an error rather than being handed to the JSON parser, which would fail
// with a message that says nothing about what went wrong.
const postJson = async <T>(url: string, body: unknown): Promise<T> => {
    const response = await fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
    });

    if (!response.ok) {
        if (response.status === 429) {
            throw new Error("Too many requests, please try again in a moment.");
        }

        throw new Error(`Request failed with status ${response.status}`);
    }

    return (await response.json()) as T;
};

export const fetchWeather = (location: Location): Promise<WeatherData> =>
    postJson<WeatherData>("/data/weather", location);

export const fetchPois = async (
    category: string,
    location: Location,
): Promise<poiElement[]> => {
    const pois = await postJson<poiElement[] | null>("/data/poi", {
        category,
        ...location,
    });

    return pois ?? [];
};
