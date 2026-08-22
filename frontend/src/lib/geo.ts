import type { Location } from "./api";

const earthRadius = 6371; // km

const toRadians = (degrees: number): number => (degrees / 180) * Math.PI;

// Haversine distance in kilometres, matching the calculation the server uses.
export const distance = (from: Location, to: Location): number => {
    const lat1 = toRadians(from.lat);
    const lon1 = toRadians(from.lon);
    const lat2 = toRadians(to.lat);
    const lon2 = toRadians(to.lon);

    return (
        2 *
        earthRadius *
        Math.asin(
            Math.sqrt(
                (1 -
                    Math.cos(lat2 - lat1) +
                    Math.cos(lat1) * Math.cos(lat2) * (1 - Math.cos(lon2 - lon1))) /
                    2,
            ),
        )
    );
};
