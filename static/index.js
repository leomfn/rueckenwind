
const addPoiHandlers = (body) => {
    document.getElementById('sites-fab-container').style.visibility = 'visible';

    const sitesFabMain = document.getElementById('sites-fab-main');
    const sitesFabChoices = document.getElementsByClassName('sites-fab-choices');

    sitesFabMain.addEventListener('click', () => {
        Array.from(sitesFabChoices).forEach(element => {
            element.classList.toggle('collapsed');
        });
    })

    const campingButton = document.getElementById('sites-fab-camping');
    const waterButton = document.getElementById('sites-fab-water');
    const cafeButton = document.getElementById('sites-fab-cafe');

    const poiLoadingIndicator = document.createElement('div');
    poiLoadingIndicator.id = 'poi-loader';

    const campingIcon = document.createElement('img');
    campingIcon.src = '/static/images/campsite.svg';

    const waterIcon = document.createElement('img');
    waterIcon.src = '/static/images/water.svg';

    const cafeIcon = document.createElement('img');
    cafeIcon.src = '/static/images/coffee.svg';

    const sitesFabMainImage = document.getElementById('sites-fab-main-image');

    campingButton.addEventListener('click', () => {
        if (!campingButton.className.includes('sites-fab-selected')) {
            Array.from(document.getElementsByClassName('sites-container')).forEach(element => {
                element.classList.add('hidden');
            })
            body.category = 'camping'

            sitesFabMainImage.style.visibility = 'hidden';

            // If poi category has been loaded before, just show it instead of
            // loading it from the server
            const campingPois = document.getElementById('camping-pois');
            if (campingPois != null) {
                campingPois.classList.toggle('hidden');

                campingButton.classList.add('sites-fab-selected');

                waterButton.classList.remove('sites-fab-selected');
                cafeButton.classList.remove('sites-fab-selected');

                sitesFabMainImage.src = '/static/images/campsite.svg';
                sitesFabMainImage.style.visibility = 'visible';
            } else {
                htmx.ajax(
                    'POST',
                    '/data/sites',
                    {
                        target: '#compass',
                        swap: 'afterbegin',
                        values: body,
                        indicator: '#poi-loader',
                    }
                )
                    .then(() => {
                        campingButton.classList.add('sites-fab-selected');

                        waterButton.classList.remove('sites-fab-selected');
                        cafeButton.classList.remove('sites-fab-selected');

                        sitesFabMainImage.src = '/static/images/campsite.svg';
                        sitesFabMainImage.style.visibility = 'visible';
                    })
            }
        }
        sitesFabMain.click();
    })

    waterButton.addEventListener('click', () => {
        if (!waterButton.className.includes('sites-fab-selected')) {
            Array.from(document.getElementsByClassName('sites-container')).forEach(element => {
                element.classList.add('hidden');
            })
            body.category = 'drinking-water'

            sitesFabMainImage.style.visibility = 'hidden';

            // If poi category has been loaded before, just show it instead of
            // loading it from the server
            const drinkingWaterPois = document.getElementById('drinking-water-pois');
            if (drinkingWaterPois != null) {
                drinkingWaterPois.classList.toggle('hidden');

                // TODO: Find solution for code duplication
                waterButton.classList.add('sites-fab-selected');

                campingButton.classList.remove('sites-fab-selected');
                cafeButton.classList.remove('sites-fab-selected');

                sitesFabMainImage.src = '/static/images/water.svg';
                sitesFabMainImage.style.visibility = 'visible';
            } else {
                htmx.ajax(
                    'POST',
                    '/data/sites',
                    {
                        target: '#compass',
                        swap: 'afterbegin',
                        values: body,
                    }
                )
                    .then(() => {
                        waterButton.classList.add('sites-fab-selected');

                        campingButton.classList.remove('sites-fab-selected');
                        cafeButton.classList.remove('sites-fab-selected');

                        sitesFabMainImage.src = '/static/images/water.svg';
                        sitesFabMainImage.style.visibility = 'visible';
                    })
            }
        }
        sitesFabMain.click();
    })

    cafeButton.addEventListener('click', () => {
        if (!cafeButton.className.includes('sites-fab-selected')) {
            Array.from(document.getElementsByClassName('sites-container')).forEach(element => {
                element.classList.add('hidden');
            })
            body.category = 'cafe'

            sitesFabMainImage.style.visibility = 'hidden';

            // If poi category has been loaded before, just show it instead of
            // loading it from the server
            const cafePois = document.getElementById('cafe-pois');
            if (cafePois != null) {
                cafePois.classList.toggle('hidden');

                cafeButton.classList.add('sites-fab-selected');

                waterButton.classList.remove('sites-fab-selected');
                campingButton.classList.remove('sites-fab-selected');

                sitesFabMainImage.src = '/static/images/coffee.svg';
                sitesFabMainImage.style.visibility = 'visible';
            } else {
                htmx.ajax(
                    'POST',
                    '/data/sites',
                    {
                        target: '#compass',
                        swap: 'afterbegin',
                        values: body,
                    }
                )
                    .then(() => {
                        cafeButton.classList.add('sites-fab-selected');

                        waterButton.classList.remove('sites-fab-selected');
                        campingButton.classList.remove('sites-fab-selected');

                        sitesFabMainImage.src = '/static/images/coffee.svg';
                        sitesFabMainImage.style.visibility = 'visible';
                    })
            }
        }
        sitesFabMain.click();
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

const renderCompassAndInfo = (position) => {
    const body = {
        lat: position.lat,
        lon: position.lon,
    }

    htmx.ajax(
        'POST',
        '/data/weather',
        {
            target: '#welcome-info',
            swap: 'outerHTML',
            values: body,
        }
    )
        .then(() => {
            addCompassRotation()
        })
        .then(() => {
            addPoiHandlers(body)
        })
}

getPosition();
