type poiElement = {
    bearing: number;
    distance: number;
    distance_pixel: number;
    distance_text: string;
    name: string;
    website: string;
}

export type Pois = {
    [key: string]: poiElement[];
};
