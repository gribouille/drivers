'use strict';

const URI = `${window.location.protocol}//${window.location.host}/events`
const PARIS = [48.866667, 2.333333]
const ZOOM = 12
const MAPBOX = 'https://api.tiles.mapbox.com/v4/{id}/{z}/{x}/{y}.png?access_token={accessToken}'
const TOKEN = 'pk.eyJ1IjoiZ3JpYm91aWxsbGUiLCJhIjoiY2o5ZTFwemFuMjUxdzJ3b3JvZWJ3Ym5wdyJ9.PXmGJAj82SZOCt_oupYLhA'
const RADIUS = 5
const STATE_COLORS = {
  'ride': 'red',
  'available': 'green'
}

const drivers = {}
const evtSource = new EventSource(URI)
const map = L.map('main-map').setView(PARIS, ZOOM)

L.tileLayer(MAPBOX, {
    attribution: 'Drivers',
    maxZoom: 18,
    id: 'mapbox.streets',
    accessToken: TOKEN
}).addTo(map)

function updateMap(driver) {
  // console.log(driver)
  const old = drivers[driver.id]
  if (old !== undefined && old !== null) {
    old.remove()
  }
  const color = STATE_COLORS[driver.state];
  drivers[driver.id] = L.circle(driver.position, {
      color: color,
      fillColor: color,
      fillOpacity: 0.5,
      radius: RADIUS
  }).addTo(map)
}

evtSource.onmessage = function(e) {
  const driver = JSON.parse(e.data)
  updateMap(driver)
}

evtSource.onerror = function(e) {
  console.error(e)
}

evtSource.onopen = function(e) {
  console.log("Connection to server")
}
