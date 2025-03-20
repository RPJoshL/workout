/* Shared leaflet code */ import { GetLeafletMap } from "internal/api/components/leaflet/lefMap";
import L from "leaflet";

interface Options {
	textStart: string;
	textEnd: string;
	markerPopupContent: string;

	startLocation: CircleArea | null
	endLocation: CircleArea | null

	lastWorkoutLocation: {
		latitude: number
		longitude: number
	} | null
}

interface CircleArea {
	radius: number;
	center: {
		latitude: number,
		longitude: number;
	}
}

/** Global variable to save the layergroups in */
const groups: Record<string, L.LayerGroup> = {
	start: L.layerGroup(),
	end: L.layerGroup()
}
let map: L.Map

export function InitMap(id: string, options: Options) {
	groups.start.clearLayers()
	groups.end.clearLayers()

	map = GetLeafletMap(id, getMapOptions(options))
	const baseLayer = L.featureGroup()
	baseLayer.addLayer(groups.start)
	baseLayer.addLayer(groups.end)
	map.addLayer(baseLayer)

	// Add initial markes
	createMarkerByCircle("start", options.startLocation, options)
	createMarkerByCircle("end", options.endLocation, options)

	// Center map for layer groups
	if (options.startLocation !== null || options.endLocation !== null) {
		// Cannot use bound sof feature group (see https://github.com/Leaflet/Leaflet/issues/7228).
		// Crating bounds manually
		const bounds: L.LatLng[] = []
		let maxRadius = 0
		if (options.startLocation !== null) {
			bounds.push(L.latLng(options.startLocation.center.latitude, options.startLocation.center.longitude))
			if (options.startLocation.radius > maxRadius) maxRadius = options.startLocation.radius
		} 
		if (options.endLocation !== null) {
			bounds.push(L.latLng(options.endLocation.center.latitude, options.endLocation.center.longitude))
			if (options.endLocation.radius > maxRadius) maxRadius = options.endLocation.radius
		}
		
		const bound = bounds[0].toBounds(maxRadius * 1.5)
		if (bounds.length >= 2) bound.extend(bounds[1].toBounds(maxRadius * 1.5))

		map.fitBounds(bound)
	} else if (options.lastWorkoutLocation !== null) {
		const l = options.lastWorkoutLocation
		map.fitBounds(new L.LatLng(l.latitude, l.longitude, 0).toBounds(1000))
	}
}

function createMarkerByCircle(context: "start" | "end", a: CircleArea | null, options: Options) {
	if (a === null) return

	createMarker(
		context, 
		context === "start" ? options.textStart : options.textEnd,
		new L.LatLng(a.center.latitude, a.center.longitude, 0),
		options.markerPopupContent,
		a.radius
	)
}

/**
 * Returns a list of leaflet options to apply to the map
 */
function getMapOptions(options: Options): L.MapOptions {
	return {
		contextmenu: true,
		contextmenuItems: [
			{
				text: options.textStart,
				callback: (e) => {
					createMarker(
						"start", options.textStart, 
						e.latlng, options.markerPopupContent,
						200
					)
				}
			},
			{
				text: options.textEnd,
				callback: (e) => {
					createMarker(
						"end", options.textEnd, 
						e.latlng, options.markerPopupContent,
						200
					)
				}
			}
		]
	}
}

function createMarker(context: "start" | "end", title: string, position: L.LatLng, popupContent: string, radius: number) {
	// Only a single marker can be set for the start / end position
	groups[context]?.clearLayers()

	// Get the input div to store the values in
	const inputDiv = window.document.getElementById("leaflet-location-"+context)!
	const inputPrefix = context === "start" ? "startLocation" : "endLocation"
	inputDiv.innerHTML = ""

	// Input fields
	const inputs = {
		radius: createInputElement(inputPrefix+".radius", radius.toString(), inputDiv),
		lat: createInputElement(inputPrefix+".center.latitude", position.lat.toString(), inputDiv),
		lon: createInputElement(inputPrefix+".center.longitude", position.lng.toString(), inputDiv)
	}

	const color = context === "start" ? "blue" : "red"
	const circle = new L.Circle(position, {
		radius: 200,
		fillColor: color,
		fillOpacity: 0.2,
		color: color,
		opacity: 0.7,
	})
	groups[context]?.addLayer(circle)

	const marker = new L.Marker(position, {
		title: title,
		icon: L.icon({
			iconUrl: '/static/img/svg/marker.svg',
			iconSize: [28, 41]
		}),

		contextmenu: true,
		contextmenuItems: [
			{
				text: 'Remove marker',
				callback: () => {
					groups[context].removeLayer(marker)
					groups[context].removeLayer(circle)
					inputDiv.innerHTML = ""
				}
			},
		],
		draggable: true,
	})
	marker.bindPopup(`${popupContent.replace("##title##", title)}`, {
		keepInView: true,
	})

	// Show popup to change radius on click
	let value = radius.toString()
	marker.on('popupopen', () => {
		const input = marker.getPopup()?.getElement()?.querySelector("input") as HTMLInputElement | null
		if (input) {
			input.setAttribute("value", value.toString())
			input.addEventListener("change", () => {
				value = input.value
				circle.setRadius(Number(value))
				inputs.radius.setAttribute("value", value)
			})

			input.addEventListener("keydown", (e) => {
				if (e.key === "Enter") {
					// Change event isn't triggered
					value = input.value
					circle.setRadius(Number(value))
					inputs.radius.setAttribute("value", value)

					e.preventDefault()
					marker.closePopup()
				}
			})
		}
	})
	marker.on("drag", () => {
		const latLng = marker.getLatLng()
		circle.setLatLng(latLng)
		inputs.lat.setAttribute("value", latLng.lat.toString())
		inputs.lon.setAttribute("value", latLng.lng.toString())
	})


	groups[context]?.addLayer(marker)
}

function createInputElement(name: string, value: string, root: HTMLElement): HTMLInputElement {
	const input = document.createElement("input")
	input.setAttribute("name", name)
	input.setAttribute("value", value)
	root.appendChild(input)

	return input
}