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
            .then(() => {
                document.getElementById('sites-fab-container').style.visibility = 'visible';

                const sitesFabMain = document.getElementById('sites-fab-main');
                const sitesFabChoices = document.getElementsByClassName('sites-fab-choices');

                sitesFabMain.addEventListener('click', () => {
                    if (sitesFabMain.className.includes('sites-fab-selected')) {
                        Array.from(sitesFabChoices).forEach(element => {
                            element.classList.remove('collapsed');
                            element.style.pointerEvents = 'auto';
                            sitesFabMain.classList.remove('sites-fab-selected');
                        });
                    } else {
                        Array.from(sitesFabChoices).forEach(element => {
                            element.classList.add('collapsed');
                            element.style.pointerEvents = 'none';
                            sitesFabMain.classList.add('sites-fab-selected');
                        });
                    }
                })

                const campingButton = document.getElementById('sites-fab-camping');
                const waterButton = document.getElementById('sites-fab-water');
                const cafeButton = document.getElementById('sites-fab-cafe');
                const campingSites = document.getElementById('camping-sites');
                const drinkingWaterSites = document.getElementById('drinking-water-sites');
                const cafeSites = document.getElementById('cafe-sites');
                
                const campingIcon = document.createElement('img');
                campingIcon.src = '/static/images/campsite.svg';

                const waterIcon = document.createElement('img');
                waterIcon.src = '/static/images/water.svg';

                const cafeIcon = document.createElement('img');
                cafeIcon.src = '/static/images/coffee.svg';

                campingButton.addEventListener('click', () => {
                    if (!campingButton.className.includes('sites-fab-selected')) {
                        campingSites.classList.add('visible');
                        campingSites.classList.remove('hidden');

                        drinkingWaterSites.classList.add('hidden');
                        drinkingWaterSites.classList.remove('visible');
                        waterButton.classList.remove('sites-fab-selected');

                        cafeSites.classList.add('hidden');
                        cafeSites.classList.remove('visible');
                        cafeButton.classList.remove('sites-fab-selected');

                        campingButton.classList.add('sites-fab-selected');
                        sitesFabMain.innerHTML = "";
                        sitesFabMain.appendChild(campingIcon);
                    }
                    sitesFabMain.click();
                })
                waterButton.addEventListener('click', () => {
                    if (!waterButton.className.includes('sites-fab-selected')) {
                        drinkingWaterSites.style.visibility = 'visible';
                        drinkingWaterSites.classList.add('visible');
                        drinkingWaterSites.classList.remove('hidden');

                        campingSites.classList.add('hidden');
                        campingSites.classList.remove('visible');
                        campingButton.classList.remove('sites-fab-selected');

                        cafeSites.classList.add('hidden');
                        cafeSites.classList.remove('visible');
                        cafeButton.classList.remove('sites-fab-selected');

                        waterButton.classList.add('sites-fab-selected');
                        sitesFabMain.innerHTML = "";
                        sitesFabMain.appendChild(waterIcon);
                    }
                    sitesFabMain.click();
                })
                cafeButton.addEventListener('click', () => {
                    if (!cafeButton.className.includes('sites-fab-selected')) {
                        cafeSites.style.visibility = 'visible';
                        cafeSites.classList.add('visible');
                        cafeSites.classList.remove('hidden');

                        drinkingWaterSites.classList.add('hidden');
                        drinkingWaterSites.classList.remove('visible');
                        waterButton.classList.remove('sites-fab-selected');

                        campingSites.classList.add('hidden');
                        campingSites.classList.remove('visible');
                        campingButton.classList.remove('sites-fab-selected');

                        cafeButton.classList.add('sites-fab-selected');
                        sitesFabMain.innerHTML = "";
                        sitesFabMain.appendChild(cafeIcon);
                    }
                    sitesFabMain.click();
                })
            })
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
