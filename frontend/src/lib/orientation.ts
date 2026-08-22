// Device orientation handling.
//
// The compass only means anything if the reading is referenced to true north.
// Chromium-based browsers provide that through 'deviceorientationabsolute',
// Firefox through 'deviceorientation' with absolute set, and iOS through the
// vendor-prefixed webkitCompassHeading, which additionally requires permission
// to be requested from a user gesture. A relative reading is deliberately
// ignored: rotating the compass by an arbitrary reference would point the wind
// arrows somewhere plausible but wrong.

export type CompassStatus =
    // Waiting for the first usable reading.
    | "pending"
    // The user has to allow access before any reading arrives (iOS).
    | "needs-permission"
    // Readings are coming in.
    | "active"
    // No usable reading is available on this device or browser.
    | "unavailable";

// iOS exposes the compass heading on a non-standard property.
interface WebkitDeviceOrientationEvent extends DeviceOrientationEvent {
    webkitCompassHeading?: number;
}

type PermissionState = "granted" | "denied";
type PermissionRequest = () => Promise<PermissionState>;

// How long to wait for a first reading before treating the compass as
// unavailable. Devices without a magnetometer stay silent instead of reporting
// an error.
const firstReadingTimeout = 3000;

// Returns iOS' permission request function, or null on browsers that do not
// gate device orientation behind a permission prompt.
const permissionRequest = (): PermissionRequest | null => {
    const orientationEvent = window.DeviceOrientationEvent as unknown as
        | { requestPermission?: PermissionRequest }
        | undefined;

    if (typeof orientationEvent?.requestPermission !== "function") {
        return null;
    }

    return orientationEvent.requestPermission.bind(orientationEvent);
};

export const needsPermission = (): boolean => permissionRequest() !== null;

// Asks for orientation access. Must be called from a user gesture, otherwise
// iOS rejects the request.
export const requestPermission = async (): Promise<boolean> => {
    const request = permissionRequest();

    if (request === null) {
        return true;
    }

    try {
        return (await request()) === "granted";
    } catch {
        return false;
    }
};

// Converts an orientation event into the rotation to apply to the compass rose,
// or null when the reading has no north reference.
const rotationFromEvent = (event: DeviceOrientationEvent): number | null => {
    const compassHeading = (event as WebkitDeviceOrientationEvent)
        .webkitCompassHeading;

    // iOS reports the heading directly, measured clockwise from north. The rose
    // has to turn against it to keep pointing north.
    if (typeof compassHeading === "number" && !Number.isNaN(compassHeading)) {
        return -compassHeading;
    }

    // Elsewhere alpha is measured counter-clockwise from north, so it is
    // already the rotation the rose needs, but only when the reading is
    // absolute.
    if (event.absolute && event.alpha !== null) {
        return event.alpha;
    }

    return null;
};

// Starts listening for orientation readings. Returns a function that stops
// listening again.
export const startCompass = (
    onRotation: (degrees: number) => void,
    onStatus: (status: CompassStatus) => void,
): (() => void) => {
    if (!window.DeviceOrientationEvent) {
        onStatus("unavailable");
        return () => {};
    }

    let receivedReading = false;

    const handleEvent = (event: Event) => {
        const rotation = rotationFromEvent(event as DeviceOrientationEvent);

        if (rotation === null) {
            return;
        }

        if (!receivedReading) {
            receivedReading = true;
            onStatus("active");
        }

        onRotation(rotation);
    };

    // Both events are registered because support differs between browsers, and
    // readings without a north reference are filtered out anyway.
    window.addEventListener("deviceorientationabsolute", handleEvent);
    window.addEventListener("deviceorientation", handleEvent);

    // Without this the compass would keep pointing north as though that were a
    // real reading, which is worse than admitting it does not work.
    const timeout = window.setTimeout(() => {
        if (!receivedReading) {
            onStatus("unavailable");
        }
    }, firstReadingTimeout);

    return () => {
        window.clearTimeout(timeout);
        window.removeEventListener("deviceorientationabsolute", handleEvent);
        window.removeEventListener("deviceorientation", handleEvent);
    };
};
