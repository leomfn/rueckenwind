const openModal = (modalIdentifier) => {
    document.getElementById(`${modalIdentifier}-modal`).style.visibility = 'visible';
    document.getElementById('site-info').style.visibility = 'hidden';
}

const closeModal = (modalIdentifier) => {
    document.getElementById(`${modalIdentifier}-modal`).style.visibility = 'hidden';
    document.getElementById('site-info').style.visibility = 'visible';
}

const renderCompassAndInfo = (position) => {
    htmx.ajax('GET', `/weather?lat=${position.lat}&lon=${position.lon}`, '#welcome-info');
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
        openModal('location-error');
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




// navigator.geolocation.getCurrentPosition(
//     position => {
//         console.log(position);
//         // const request = new Request(`/weather?lat=${position.coords.latitude}&lon=${position.coords.longitude}`);
//         // fetch(request)
//         //     .then(res => res.text())
//         //     .then(snippet => {
//         //         console.log(snippet);
//         //         htmx.ajax
//         //     })
//         htmx.ajax('GET', `/weather?lat=${position.coords.latitude}&lon=${position.coords.longitude}`, '#weather');
//     }
// )

// navigator.geolocation.getCurrentPosition(
//     position => {
//         const request = new Request(`/weather?lat=${position.coords.latitude}&lon=${position.coords.longitude}`);
//         fetch(request)
//             .then(res => res.json())
//             .then(data => {
//                 document.getElementById("currentTemp").innerText = data.temp_current;
//                 document.getElementById("futureTemp").innerText = data.temp_future;
//                 document.getElementById("currentWindSpeed").innerText = data.wind_current;
//                 document.getElementById("futureWindSpeed").innerText = data.wind_future;
//                 document.getElementById("currentWindGust").innerText = data.wind_gust_current;
//                 document.getElementById("futureWindGust").innerText = data.wind_gust_future;
//                 document.getElementById("sunsetTime").innerText = data.sunset;

//                 const currentWindArrow = document.getElementById("currentWindArrow")
//                 currentWindArrow.innerHTML = arrowSvg
//                 currentWindArrow.style = `transform: rotate(${data.wind_deg_current}deg) scale(${(data.wind_current / 120)**0.2})`;
//                 const futureWindArrow = document.getElementById("futureWindArrow")
//                 futureWindArrow.innerHTML = arrowSvg
//                 futureWindArrow.style = `transform: rotate(${data.wind_deg_future}deg) scale(${(data.wind_future / 120)**0.2})`;

//                 document.getElementById("currentRain").innerText = rainSymbol(data.rain_current);
//                 document.getElementById("futureRain").innerText = rainSymbol(data.rain_future);
//             })
//             .catch(error => console.warn("Something went wrong when trying to handle the weather data."))
        
//         // Get nearest campsites from overpass api
//         const overpassRequest = new Request("https://overpass-api.de/api/interpreter", {
//             method: "POST",
//             body: `[out:json];nwr["tourism"="camp_site"]["tent"!="no"](around:25000.0,${position.coords.latitude},${position.coords.longitude});out geom;`,
//         })
//         fetch(overpassRequest)
//             .then(res => res.json())
//             .then(data => {
//                 const sites = data.elements.map(el => {
//                     if (el.type === "node") {
//                         return {
//                             direction: bearing(position.coords.latitude, position.coords.longitude, el.lat, el.lon),
//                             distance: haversineDistance(position.coords.latitude, position.coords.longitude, el.lat, el.lon),
//                         }
//                     } else if (el.type === "way" || el.type === "relation") {
//                         const lat = (el.bounds.maxlat + el.bounds.minlat) / 2
//                         const lon = (el.bounds.maxlon + el.bounds.minlon) / 2

//                         return {
//                             direction: bearing(position.coords.latitude, position.coords.longitude, lat, lon),
//                             distance: haversineDistance(position.coords.latitude, position.coords.longitude, lat, lon),
//                         }
//                     }
//                 })

//                 sites.forEach(site => {
//                     const newSite = document.createElement("div");
//                     newSite.innerHTML = `${Math.round(site.distance)}<br>*`;
//                     newSite.className = "compassSite";
//                     newSite.style = `transform: rotate(${site.direction}deg) translateY(-20px);`;
//                     compass.appendChild(newSite);
//                 })
//             })
//     },
//     locationError,
//     locationOptions
// )