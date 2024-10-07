const renderCompassAndInfo = (position) => {
    const body = {
        lat: position.lat, 
        lon: position.lon,
    }

    htmx.ajax(
        'POST',
        '/weather',
        {
            target: '#welcome-info',
            swap: 'outerHTML',
            values: body,
        }
    )
        .then(() => {
            addCompassRotation()

            htmx.ajax(
                'POST',
                '/sites',
                {
                    target: '#sites-loader',
                    swap: 'outerHTML',
                    values: body,
                }
            )
        })
}

const addRegularOrientationEventListener = () => {
    window.addEventListener("deviceorientationabsolute", event => {
        document.getElementById('compass').style = `transform: rotate(${event.alpha}deg)`;
    }, true);
}

const addIosOrientationEventListener = () => {
    window.addEventListener("deviceorientation", event => {
        document.getElementById('compass').style = `transform: rotate(${-event.webkitCompassHeading}deg)`;
    })
}

const compassClickHandler = () => {
    DeviceOrientationEvent.requestPermission()
        .then(response => {
            if (response === "granted") {
                addIosOrientationEventListener();
            } else {
                console.warn("Could not get permissions for iPhone's sensors. Compass rotation won't work.")
            }
        })
        .catch(() => console.warn("An error occured when trying to request the sensor permissions."))
}

const addCompassRotation = () => {
    if (typeof DeviceOrientationEvent.requestPermission === "function") {
        // iOS 13 or higher
        DeviceOrientationEvent.requestPermission()
            .then(response => {
                if (response === "granted") {
                    // Permission has already been given.
                    addIosOrientationEventListener();
                }
            })
            .catch(() => {
                // Permission has not yet been given. Inform the user and
                // call for action. If the user clicks the compass, they
                // will be asked for sensor permissions, which should then
                // automatically enable compass rotation.
                htmx.ajax('GET', '/error?type=orientation', {
                    target: 'body',
                    swap: 'beforeend'
                })
                compass.addEventListener("click", compassClickHandler);
            })
    } else {
        // Other OS
        addRegularOrientationEventListener();
    }
}

const getPosition = () => {
    const locationSuccess = (location) => {
        position = {
            lat: location.coords.latitude,
            lon: location.coords.longitude
        };
        
        renderCompassAndInfo(position);
    }

    const locationFailure = () => {
        htmx.ajax('GET', '/error?type=location', {
            target: 'body',
            swap: 'beforeend'
        })
    }

    const locationOptions = {
        enableHighAccuracy: true
    }

    if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(
            locationSuccess,
            locationFailure,
            locationOptions
        );
    }
}

getPosition();
