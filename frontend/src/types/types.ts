export type poiElement = {
    bearing: number;
    distance: number;
    distance_pixel: number;
    distance_text: string;
    address: string;
    name: string;
    website: string;
    lon: number;
    lat: number;
}

export type Pois = {
    [key: string]: poiElement[];
};
