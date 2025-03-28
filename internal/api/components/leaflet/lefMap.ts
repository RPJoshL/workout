import L from "leaflet";
import { BingLayer, NavigationControl, SmoothPoly } from "./extension";

/**
 * Returns a list of layers 
 */
function GetBaseLayers(): L.Control.LayersObject {
	const highDpi = window.devicePixelRatio > 1;
	const osmMapnik = new L.TileLayer(
		highDpi ? 'https://tile.osmand.net/hd/{z}/{x}/{y}.png' : 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
		{
			minZoom: 3,
			maxZoom: 20,
			maxNativeZoom: 19,
			attribution: '©OpenStreetMap',
			detectRetina: false
		}
	);

	const osmOpenTopoMap = new L.TileLayer(
		'https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png',
		{
			minZoom: 3,
			maxZoom: 20,
			maxNativeZoom: 17,
			attribution: '©<a href="https://openstreetmap.org/copyright">OpenStreetMap</a>-Mitwirkende, ©<a href="https://opentopomap.org">OpenTopoMap</a> (<a href="https://creativecommons.org/licenses/by-sa/3.0/">CC-BY-SA</a>)',
			detectRetina: false
		}
	);

	const googleMaps = new L.TileLayer(
		'https://mt.google.com/vt?&x={x}&y={y}&z={z}',
		{
			attribution: "<a href='http://maps.google.com/'>Google</a> Maps",
			tileSize: 256,
			minZoom: 3,
			maxZoom: 20,
			maxNativeZoom: 20,
			detectRetina: false
		}
	);

	const googleSatellite = new L.TileLayer(
		'https://mt.google.com/vt?lyrs=s&x={x}&y={y}&z={z}',
		{
			attribution:"<a href='http://maps.google.com/'>Google</a> Maps Satellite",
			tileSize:256,
			minZoom:3,
			maxZoom: 20,
			maxNativeZoom:20,
			detectRetina: false
		}
	);

	const googleHybrid = new L.TileLayer(
		'https://mt.google.com/vt?lyrs=y&x={x}&y={y}&z={z}',
		{
			attribution:"<a href='http://maps.google.com/'>Google</a> Maps Satellite",
			tileSize:256,
			minZoom:3,
			maxZoom: 20,
			maxNativeZoom:20,
			detectRetina: false
		}
	);


	const bingMaps = new BingLayer(
		'https://ecn.t{s}.tiles.virtualearth.net/tiles/r{q}?g=864&mkt=en-gb&lbl=l1&stl=h&shading=hill&n=z',
		{
			subdomains: "0123",
			minZoom: 3,
			maxZoom: 20,
			maxNativeZoom: 19,
			attribution: "<a href='http://maps.bing.com/'>Bing</a> map data copyright Microsoft and its suppliers",
			detectRetina: false
		}
	);
	const bingAerial = new BingLayer(
		'https://ecn.t{s}.tiles.virtualearth.net/tiles/a{q}?g=737&n=z',
		{
			subdomains: "0123",
			minZoom: 3,
			maxZoom: 20,
			maxNativeZoom: 19,
			attribution: "<a href='http://maps.bing.com/'>Bing</a> map data copyright Microsoft and its suppliers",
			detectRetina: false
		}
	);

	return {
		"OpenStreetMap": osmMapnik,
		"OpenTopoMap": osmOpenTopoMap,
		"Google Maps": googleMaps,
		"Google Maps Satellite": googleSatellite,
		"Google Maps Hybrid": googleHybrid,
		"Bing Maps": bingMaps,
		"Bing Aerial View": bingAerial,
	};
}

type HeartrateColor = {
	minHeartrate: number
	maxHeartrate: number
	color: string
}
function GetHeartrateColors(): Array<HeartrateColor> {
	return [
		{
			minHeartrate: 0,
			maxHeartrate: 97,
			color: "#ff8a80"
		},
		{
			minHeartrate: 97,
			maxHeartrate: 116,
			color: "#2862ff"
		},
		{
			minHeartrate: 116,
			maxHeartrate: 135,
			color: "#00cee9"
		},
		{
			minHeartrate: 135,
			maxHeartrate: 154,
			color: "#65dd19"
		},
		{
			minHeartrate: 154,
			maxHeartrate: 174,
			color: "#ff6d01"
		},
		{
			minHeartrate: 174,
			maxHeartrate: 200,
			color: "#aa00ff"
		},
	]
}

function GetColorByHeartrate(heartrate: number): string {
	const res = GetHeartrateColors().find(hc => heartrate >= hc.minHeartrate && heartrate < hc.maxHeartrate )
	return res?.color ?? "#12f"
}

/** Point is a single marker point passed by go  */
type DPoint = {
	Latitude: number
	Longitude: number
	Heartrate: number
	Distance: number
	TooltipContent: string
	PartIndex: number
}

/** Line is a single line to display on the map */
type Line = {
	TooltipContent: string
	PartIndex: number
	Points: Array<DPoint>
}

/**
 * AddLeaflet initializes a leaflet map into the object identified
 * by the provided ID
 * 
 * @param id 	ID of the HTML element to render the map into
 */
export function AddLeaflet(id: string, line: Array<DPoint> | null, lines: Array<Line> | null) {
	let linesControl: L.Control | undefined;
	const map = GetLeafletMap(
		id, {
			contextmenu: true,
			contextmenuItems: [
				{
					text: 'Show coordinates',
					callback: function (e) {
						showCoordinates(e.latlng)
					}
				},
				{
					text: 'Add marker',
					callback: function (e) {
						createMarker(map, e.latlng).addTo(map);
					}
				},
			],
		},
		(map, layerControl) => {
			// Display all lines on map
			displayLine(line, map, layerControl)

			// Display multiple lines on the map
			linesControl = displayLines(lines, map)
		}
	)

	// Layer control has to be added at the end to propagate checkboxes correctly
	if (linesControl !== undefined) linesControl.addTo(map)
}

/**
 * Initializes a new leaflet map instance to use in the app 
 * with some default options and returns it.
 * 
 * @param id					Unique ID of the map instance
 * @param additionalOptions		Additional leaflet options to use during map creation
 * @param configureLayers		Optional function to set your custom layers
 */
export function GetLeafletMap(
	id: string, 
	additionalOptions: L.MapOptions | null = null, 
	configureLayers: ((map: L.Map, control: L.Control.Layers) => void) | null = null
): L.Map {

	// Map sources
	const baseLayers = GetBaseLayers()
	
	// Configure map
	const map = L.map(id, {
		layers: [ baseLayers["OpenStreetMap"] ],
		maxBounds: [[90,-500], [-90,500]], // to prevent getting lost in north/south
		worldCopyJump: true,
		// Contextmenu options are not optional...
		contextmenu: false,
		contextmenuItems: [],
		dragging: !L.Browser.mobile,
		tap: !L.Browser.mobile,
	
		fullscreenControl: true,
		...additionalOptions
	});
	
	// Set main view
	map.setView([51.505, -0.09], 15);
	
	// Create a pane with a very high z-index
	map.createPane("overlay-pane")
	map.getPane("overlay-pane")!.style.zIndex = "999"
	
	// Scale unint in the left corner
	L.control.scale({ imperial: false }).addTo(map);
	
	/* Layer selector with base and overlay layers */
	const layerControl = L.control.layers(baseLayers, undefined, { collapsed: true })
	
	// Add custom layers
	if (configureLayers !== null) configureLayers(map, layerControl)
	layerControl.addTo(map);
	
	// Add resize listener for the map so that tiles are loaded correctly.
	const resizeObserver = new ResizeObserver(() => {
		map.invalidateSize();
	});
	resizeObserver.observe(document.getElementById(id)!);
	
	// Add map to global window variable
	(window as any)["leaflet-map-"+id] = map
	
	map.on("fullscreenchange", () => {
		if (map.isFullscreen()) {
			map.dragging.enable()
			map.tap?.enable()
		} else {
			if (L.Browser.mobile) {
				map.dragging.disable()
				map.tap?.disable()
			}
		}
	})

	return map
}

/** TrackElements contains all layer groups that are required 
 * for displaying a single track segment
 */
type TrackElements = {
	/** Markers every 1 km */
	marker: Array<{ count: number, marker: L.Marker }>

	/** Colored lines */
	color: Array<SmoothPoly>
	/** Polylines for displaying a border around colored polyline to improve contrast */
	border: Array<SmoothPoly>

	/** Static polyline rendered in a single color */
	static: SmoothPoly

	/** All points that do belong to this track */
	points: Array<L.LatLngExpression>

	/** Polylines which do contain the popover */
	popover: Array<L.Polyline>

	/** Weather the track is currently visible */
	isVisible: boolean
}

/**
 * Displays a single connected line defined by the provided points
 */
function displayLine(points: Array<DPoint> | null, map: L.Map, control: L.Control.Layers) {
	// Nothing to do
	if (points == null || points.length < 2) return

	const polyLineProperties: L.PolylineOptions = {
		weight: 5,
		smoothFactor: 2,
		stroke: true,
		className: "stroke-polyline",
		interactive: false,
		color: points[0].Heartrate > 20 ? GetColorByHeartrate(points[0].Heartrate) : "#12f"
	};
	const borderPolylineProperties: L.PolylineOptions = {
		color: "#000",
		weight: (polyLineProperties.weight ?? 4) + 2,
		smoothFactor: 2,
		stroke: true,
		interactive: false
	}

	// Group on which all points are displayed
	const group = L.featureGroup();
	control.addOverlay(group, "Km markers and popups")
	const groupColor = L.featureGroup()
	const groupBorder = L.featureGroup()
	control.addOverlay(groupColor, "Colored lines")
		
	// Parse and show all points on the group
	let prevPoint: DPoint;

	/** Ignore color change checks if a color change should be ignored if the last color is only "temporary" */
	const ignoreColorChange = (startIndex: number, currentColor: string, pt: DPoint): boolean => {
		const maxAllowedOffset = 180
		let otherColorFound = false
		for (let ii = startIndex; ii < points.length; ii++) {
			const distance = points[ii].Distance - pt.Distance
			if (distance > maxAllowedOffset * 2) {
				break
			}
			
			// Same color found again
			const nextColor = GetColorByHeartrate(points[ii].Heartrate)
			if (nextColor == currentColor && (distance < maxAllowedOffset || ii < startIndex + 1)) {
				otherColorFound = true
			}
		}

		return otherColorFound
	}

	/** Displays the km markers based on the zoom level and the tracks visiblity */
	const displayKmMarker = (zoom: number, segmentIndex: string | null = null) => {
		// Adjust km points (remove all and add again)
		let everyKm = 1
		if (zoom < 11) everyKm = -1
		else if (zoom < 12) everyKm = 4
		else if (zoom < 13) everyKm = 2
		else everyKm = 1

		Object.keys(segments).forEach(segment => {
			const seg = segments[segment]

			if (seg.isVisible && (segmentIndex === null || segmentIndex === segment)) {
				seg.marker.forEach(e => group.removeLayer(e.marker))
				seg.marker.filter(m => everyKm != -1 && m.count % everyKm === 0).forEach(m => m.marker.addTo(group))
			}
		})
	}

	/** Map indexed by the track segment with all the elements that are displayed on the leaflet map */
	const segments: Record<string, TrackElements> = {}
	let kmCounter = 0

	// Loop over each point
	let lastPoints: Array<L.LatLngExpression> = []
	/** Index to which color comparisons are ignored (excluding) */
	let ignoreIndex = 0
	points.forEach((pt, i) => {
		const p: L.LatLngExpression = [pt.Latitude, pt.Longitude];

		// Segment to use for this point
		if (segments[pt.PartIndex] === undefined) {
			segments[pt.PartIndex] = {
				border: [], color: [], marker: [],
				points: [], popover: [],
				static: new SmoothPoly([], polyLineProperties),
				isVisible: true,
			}
		}
		const segment = segments[pt.PartIndex]
		segment.points.push(p)
	
		if (prevPoint) {

			// Add tooltip with provided content
			if (pt.TooltipContent != "") {
				const popover = L.polyline([ [prevPoint.Latitude, prevPoint.Longitude], p ], {
					opacity: 0,
					fill: false,
					weight: 20
				}).addTo(group).bindTooltip(pt.TooltipContent, { sticky: true })

				segment.popover.push(popover)
			}

			// Check if a km counter has to be addedd
			const kmCount = Math.floor(pt.Distance / 1000)
			if (kmCount > kmCounter) {
				let leftKm = pt.Distance % 1000

				// Check if the last point is closer than this one
				const prevLeftKm = 1000 - prevPoint.Distance % 1000
				let kmPoint = pt
				if (prevLeftKm < leftKm) {
					leftKm = prevLeftKm
					kmPoint = prevPoint
				} 

				const icon = L.divIcon({ className: "km-marker", html: kmCount.toString(), iconSize: L.point(15, 15) });
				const marker = L.marker(L.latLng(kmPoint.Latitude, kmPoint.Longitude), { 
					icon: icon,
					contextmenu: false,
					contextmenuItems: []
				})
				marker.bindTooltip(pt.TooltipContent).addTo(map)

				segment.marker.push({count: kmCount, marker: marker})
				kmCounter++
			}
			
			// Render the last 
			const currentColor = GetColorByHeartrate(pt.Heartrate)
			const forceRender = points.length > (i + 1) && pt.PartIndex !== points[i+1].PartIndex
			if (forceRender || (polyLineProperties["color"] != currentColor && i >= ignoreIndex) ) {

				// The color is different. But we don't change the color for points that are
				// only present for distances < 200 meters and changes back to the default color
				// again!
				const otherColorFound = ignoreColorChange(i, polyLineProperties["color"] ?? "", pt)
				if (otherColorFound && !forceRender) {
					lastPoints.push(p)
				} else {
					// Render line
					lastPoints.push(p)
					segment.color.push(new SmoothPoly(lastPoints, polyLineProperties))
					segment.border.push(new SmoothPoly(lastPoints, borderPolylineProperties))

					// Style new line
					if (pt.Heartrate < 30 || pt.Heartrate > 200) {
						polyLineProperties["color"] = "#12f"
					} else {
						polyLineProperties["color"] = currentColor
					}
					lastPoints = [ p ]

					// If the new heart rate zone is only "valid" for a few meters and followed
					// by a "higher" zone, add the "higher" zone directly instead of displaying the
					// old one
					let zoneSmall = false
					let sameColor = true
					let newColor: string = ""
					let newIgnoreIndex = 0
					for (let ii = i+1; ii < points.length; ii++) {
						const distance = points[ii].Distance - pt.Distance 

						const nextColor = GetColorByHeartrate(points[ii].Heartrate)
						if (!zoneSmall && nextColor != currentColor && distance < 150) {
							zoneSmall = true
							newColor = nextColor
							newIgnoreIndex = ii
						}

						if (zoneSmall && newColor != nextColor && !ignoreColorChange(ii, newColor, points[ii]) && distance < 200) {
							sameColor = false
						}

						if (distance > 500) {
							break
						}
					}
					
					if (zoneSmall && sameColor && !forceRender) {
						polyLineProperties["color"] = newColor
						ignoreIndex = newIgnoreIndex
					}
				}
			} else {
				// Only add point to polyline
				lastPoints.push(p)
			}
		} else {
			lastPoints.push(p)
		}

		prevPoint = pt
	});

	// Add remaining last lines
	if (lastPoints.length > 2) {
		const segment = segments[points[points.length-1].PartIndex]

		segment.color.push(new SmoothPoly(lastPoints, polyLineProperties))
		segment.border.push(new SmoothPoly(lastPoints, borderPolylineProperties))
	}

	// Render all points
	Object.keys(segments).forEach(segment => {
		const seg = segments[segment]

		seg.static = new SmoothPoly(seg.points, {
			...polyLineProperties,
			color: "#12f",
		})
		seg.marker.forEach(m => group.addLayer(m.marker))
		seg.color.forEach(c => c.addTo(groupColor))
		seg.border.forEach(border => border.addTo(groupBorder))
	})

	// Add layer control for multiple tracks
	if (Object.keys(segments).length > 1) {
		Object.keys(segments).forEach(segment => {
			const dummyGroup = L.featureGroup()
			control.addOverlay(dummyGroup, "Segment " + segment)
			map.addLayer(dummyGroup)

			const seg = segments[segment]
			dummyGroup.addEventListener("remove", () => {
				seg.isVisible = false
				seg.popover.forEach(e => group.removeLayer(e))
				seg.border.forEach(e => groupBorder.removeLayer(e))
				seg.color.forEach(e => groupColor.removeLayer(e))
				seg.marker.forEach(e => group.removeLayer(e.marker))
				
			})
			dummyGroup.addEventListener("add", () => {
				seg.isVisible = true
				seg.popover.forEach(e => e.addTo(group))
				seg.border.forEach(e => {
					e.addTo(groupBorder)
					// Bring to back to not overlay overlapping courses
					e.bringToBack()
				})
				seg.color.forEach(e => e.addTo(groupColor))
				displayKmMarker(map.getZoom(), segment)
			})
		})
	}

	// Get status of group layer
	let groupColorHidden = false;
	groupColor.addEventListener("remove", () => { 
		groupColorHidden = true
		groupBorder.removeFrom(map)
		if (lastZoom > 11) {
			Object.keys(segments).forEach( k => segments[k].static.addTo(map) )
		}
	})
	groupColor.addEventListener("add", () => { 
		groupColorHidden = false
		groupBorder.addTo(map)
		groupBorder.bringToBack()
		if (lastZoom > 11) {
			Object.keys(segments).forEach( k => segments[k].static.removeFrom(map) )
		}
	})

	// Add last point marker
	const last = points[points.length - 1]
	const firstCircle = L.circleMarker([last.Latitude, last.Longitude], {
		color: "red",
		radius: 8,
	})
	group.addLayer(firstCircle.addTo(map).bindTooltip("End"),);
		
	// Add start pointer
	const first = points[0];
	const lastCircle = L.circleMarker([first.Latitude, first.Longitude], {
		color: "blue",
		radius: 8,
	})
	group.addLayer(lastCircle.addTo(map).bindTooltip("Start"));
		
	let lastZoom = map.getZoom()
	map.on("zoomend", () => {
		const newZoom = map.getZoom()

		// Adjust lines
		if (lastZoom > 11 && newZoom <= 11) {
			// Hide all layers
			Object.keys(segments).map(s => segments[s]).filter(s => s.isVisible).forEach(seg => {
				seg.static.addTo(map) 
			})

			groupBorder.getLayers().forEach(layer => groupBorder.removeLayer(layer))
			groupColor.getLayers().forEach(layer => groupColor.removeLayer(layer))
		} else if (lastZoom <= 11 && newZoom > 11) {
			Object.keys(segments).forEach(segment => {
				const seg = segments[segment]

				if (seg.isVisible) {
					seg.border.forEach(e =>e.addTo(groupBorder))
					seg.color.forEach(e => e.addTo(groupColor))
				}

				if (!groupColorHidden) {
					seg.static.removeFrom(map)
				}
			})

			// Bring start and end to foreground
			firstCircle.bringToFront(); lastCircle.bringToFront()
		}

		displayKmMarker(newZoom)

		lastZoom = newZoom
	})

	// Show layer by default
	map.addLayer(groupBorder)
	map.addLayer(groupColor)
	map.addLayer(group)
	
	// Resize map to bound
	map.fitBounds(group.getBounds());

	// Bring start and end to foreground
	firstCircle.bringToFront(); lastCircle.bringToFront()
}

/** Displays multiple distinct lines on the map as a raw polyline without
 * any additionals features used in [displayLine]
 */
function displayLines(lines: Array<Line> | null, map: L.Map): L.Control | undefined {
	// Nothing to do
	if (lines == null || lines.length == 0) return

	/** Group to display all lines on */
	const group = L.featureGroup();
	/** Group that contains markers to start position for every workout */
	const markerGroup = L.featureGroup()

	/** The currently selected / viewed line */
	let currentWorkout = 0
	const mapLines: Array<{ circle: L.CircleMarker, line: L.Polyline }> = []

	// We don't display workouts which don't have any points
	lines = lines.filter(l => l.Points !== null)
	lines.filter(l => l.Points.length > 2).forEach( (l, i) => {

		// Get all points for Polyline
		const allPoints: L.LatLngExpression[] = l.Points.map(pt => {
			return [ pt.Latitude, pt.Longitude ] as L.LatLngExpression
		}).reverse()

		// Color of the line
		const color = '#12f'

		// Configure polyline
		const polyLineProperties: L.PolylineOptions = {
			weight: 5,
			smoothFactor: 2,
			stroke: true,
			color: color
		};

		// Render polyline with tooltip
		const polyline = new SmoothPoly(allPoints, polyLineProperties).addTo(map)
		// if (l.TooltipContent !== "") polyline.bindTooltip(l.TooltipContent, { sticky: true })

		// Add to group
		group.addLayer(polyline)
		
		// Add last point marker
		const last = l.Points[l.Points.length - 1]
		const firstCircle = L.circleMarker([last.Latitude, last.Longitude], {
			color: color,
			radius: 6,
		})
		group.addLayer(firstCircle.addTo(map).bindPopup(l.TooltipContent),);
		
		// Add start pointer
		const first = l.Points[0];
		const lastCircle = L.circleMarker([first.Latitude, first.Longitude], {
			color: color,
			radius: 6,
		})
		group.addLayer(lastCircle.addTo(map).bindPopup(l.TooltipContent));

		// @ts-expect-error No HTMX types
		firstCircle.on("popupopen", (e) => htmx.process(e.popup.getElement()))
		// @ts-expect-error No HTMX types
		lastCircle.on("popupopen", (e) => htmx.process(e.popup.getElement()))

		// Add marker
		const marker = new L.Marker(L.latLng(first.Latitude, first.Longitude) , {
			icon: L.icon({
				iconUrl: '/static/img/svg/marker.svg',
				iconSize: [28, 41]
			}),

			contextmenu: false,
			contextmenuItems: []
		})

		// marker.bindTooltip(l.TooltipContent, { offset: L.point(15, 0) })
		marker.bindPopup(l.TooltipContent, { offset: L.point(0, -15) })
		let lastOpen = 1
		marker.on("click", () => {
			setTimeout(() => {
				if ( (marker.isPopupOpen() && (Date.now() - lastOpen) > 100) || lastOpen == 0) {
					map.fitBounds(polyline.getBounds(), { paddingTopLeft: L.point(0, 0) })
					//lastCircle.openPopup()
					currentWorkout = i
				}
			}, 40)
		})
		marker.on("popupopen", (e) => {
			lastOpen = Date.now()

			// @ts-expect-error No HTMX types
			htmx.process(e.popup.getElement())
		})
		marker.on("popupclose", () => {
			lastOpen = 0
		})

		marker.addTo(markerGroup)

		// Append to array
		mapLines.push({ circle: lastCircle, line: polyline })
	})

	// Add zoom hook
	let lastZoom = map.getZoom()
	map.on("zoomend", () => {
		const newZoom = map.getZoom()

		if (lastZoom < 10 && newZoom >= 10) {
			map.addLayer(group)
			map.removeLayer(markerGroup)
		}
		if (lastZoom >= 10 && newZoom < 10) {
			map.removeLayer(group)
			map.addLayer(markerGroup)
		}
	
		lastZoom = newZoom
	})

	// Add group to map
	group.addTo(map)
	map.fitBounds(group.getBounds())

	// Display navigation control
	const goToWorkout = (next: boolean) => {

		// Get the next workout index to focus
		let nextIndex = currentWorkout + ( next ? 1 : -1 )
		if (nextIndex >= mapLines.length) nextIndex = 0
		else if (nextIndex < 0) nextIndex = mapLines.length - 1

		// Close any opened popus
		mapLines[currentWorkout].circle.closePopup()

		// Zoom to next workout and open popup
		map.fitBounds(mapLines[nextIndex].line.getBounds(), { 
			paddingTopLeft: L.point(10, 120) 
		})
		setTimeout(() => {
			mapLines[nextIndex].circle.openPopup()
		}, 250)
		
		// Update current index
		currentWorkout = nextIndex
	}

	const nav = new NavigationControl({ 
		position: "bottomright", 
		OnNext: () => goToWorkout(true),
		OnPrevious: () => goToWorkout(false) 
	})

	return nav
}

/** Executes the provided function if the user clicks on an already opened tooltip */
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function onPupupClicked(marker: L.Marker, f: () => void) {
	let tooltipLastOpened = 0
	let lastClose = 0
	marker.on("click", () => {
		setTimeout(() => {
			// Add a delay for mobile
			if (marker.isTooltipOpen() && (Date.now() - tooltipLastOpened) > 500) {
				marker.closeTooltip()
				f()
			}
		}, 50)
	})
	marker.on("popupopen", () => {
		const lastTooltipOpened = tooltipLastOpened.valueOf()
		tooltipLastOpened = Date.now()
		const lastUpdate = Date.now()

		// If tooltip is already opened und the user clicks on it a close and open is triggered. We want to IGNORE these events
		const diff = (Math.abs(lastUpdate - lastClose))
		if (diff < 20) {
			tooltipLastOpened = lastTooltipOpened
		}
	})

	marker.on("popupclose", () => { 
		lastClose = Date.now()
	})
}

/** Display lines performant with webgl */
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function displayLineWithWebgl(map: L.Map) {

	const glify = (L as any).glify

	const data = {"type":"FeatureCollection","features":[{"type":"Feature","properties":{"scalerank":2,"name":"Brahmaputra","name_alt":null,"featureclass":"River"},"geometry":{"type":"LineString","coordinates":[[82.40047977084697,30.411477362585146],[82.72273400261909,30.3650460881709],[83.76101688022743,29.833321438103667],[84.06673465366615,29.613283189404868],[84.37906701043823,29.494504909781995],[84.70721235549163,29.29635163015881],[85.4086682474215,29.27438914643477],[87.66021040237845,29.15152842866084],[87.78872968948832,29.265604152945144],[87.9841182799839,29.34203359630483],[88.6307971536844,29.350663560497566],[91.03798872270445,29.316505438752642],[92.09549523312535,29.25340851492426],[93.25005008339039,29.069802150991237],[93.60749596555328,29.10858531342629],[94.35602908730107,29.309529120393236],[94.82039350787582,29.542150580355113],[95.15060591022092,29.810118720004624],[95.26543094277343,29.828541368116674],[95.33602094928415,29.740717271436637],[95.37260786334679,29.595739040641774],[95.31690066933615,29.41768789318013],[94.8764624369125,28.967947902943962],[94.87082970579272,28.831961371367896],[95.04353234251215,28.511231594348374],[95.08792239785086,28.21724437103991],[95.39648237506563,28.002916368109368],[95.38283979694057,27.90434357351262],[95.28992557167979,27.78786489512673],[94.83150394081858,27.464163723250437],[94.3951998229783,27.042406521100318],[93.75162153521532,26.732037868755327],[92.67788862505475,26.555123602804102],[92.32225141802209,26.464431464131863],[91.68342736202996,26.214860134378256],[91.25037885942405,26.183208319599487],[90.60607710160897,26.180831203714064],[90.33405236202455,26.13476166432585],[90.11236046749241,26.02952260996345],[89.9249817240021,25.87379568125189],[89.78803917842862,25.632105007422794],[89.61264936723,24.94837413176903],[89.71419355668354,24.584623724866532],[89.74442426957427,23.858673204030282],[90.00570031124198,23.65902130787063],[90.2256868835085,23.418415839119675],[90.38727908728518,23.11982941333723],[90.50753014522837,22.780237738531184]]}},{"type":"Feature","properties":{"scalerank":2,"name":"Mekong","name_alt":null,"featureclass":"River"},"geometry":{"type":"LineString","coordinates":[[94.08400434771664,33.15585765230966],[94.44770307818683,33.1632990585597],[94.9414197123034,33.089272569301585],[95.84043460488513,32.6322720403284],[96.2393249857461,32.52587026623944],[96.94129764199889,31.96554271089866],[97.16515994668728,31.488517564412376],[97.25678226114078,31.096474310830075],[97.84625532429419,30.277738755813786],[98.28457482299385,29.66689748790769],[98.62398563028688,28.963116156524663],[99.02809533080932,27.516279405216792],[99.15015506388275,27.027859605410157],[99.1725826354974,26.02644786224208],[99.24027876179977,25.66825267181096],[99.39127729695662,25.344758205663837],[99.6668160339369,25.013563951105226],[99.91656823120351,24.833833319594092],[100.30031741740297,24.72146291757541],[100.42129194539825,24.647255560804282],[100.48733442586726,24.52594513599911],[100.51994225464341,24.377917995699036],[100.4755521993047,24.163977566010672],[100.13851850789712,23.473115342700623],[100.12239546102205,23.311135565681738],[100.17148807169932,23.106574408454364],[100.98373823446298,21.844971828696714],[101.01825809123395,21.79771373136579],[101.15003299357824,21.849984442629022],[101.18000532430753,21.43657298429403],[100.32910119018953,20.786121731036232],[100.11598758341785,20.41784963630819],[100.54888105672688,20.109237982661128],[100.55908715210452,20.00237112068153],[100.56929324748216,19.89553009691808],[100.59730187378435,19.89304962816807],[101.11478966675517,19.85214773200906],[101.47156375529826,19.870131130446623],[102.18480187379072,20.04890574796036],[101.88373497925826,19.55518911384381],[101.76482750855465,18.721777452056614],[101.49228600464733,18.16002879482781],[101.48598147990771,17.969704494696842],[101.5636511576424,17.82051463467019],[101.58597537639247,17.810463568589427],[102.1135917500925,18.109101670804165],[102.41300499879162,17.932781683824288],[102.99870568238771,17.9616946476916],[103.20019209189374,18.309632066312773],[103.9564766784853,18.24095408779688],[104.7169470560925,17.42885895433008],[104.7793205098688,16.44186493577145],[105.58903852745016,15.570316066952856],[105.5760677429449,15.324646307837298],[105.78168826703427,15.120524400284395],[105.89103559776387,14.771088365126744],[105.86969323122733,14.217840481009958],[105.96493289594125,13.857397365774133],[105.9493266133891,13.350089829964816],[105.97227094932668,12.659641018113092],[105.92989627484735,12.397073065638082],[105.81016198122708,12.286769720911082],[105.65068851117437,12.233956407108792],[105.3681217794024,11.970484117068665],[105.04984663291677,11.804602769411758],[104.95114464723929,11.590920721884856],[105.11258182171909,10.93261465091868],[105.20797651572988,10.706866156451639],[105.5666109558355,10.26707387943165],[106.36108442589074,9.433791408725185],[90.32537072139951,47.650166734398894]]}},{"type":"Feature","properties":{"scalerank":2,"name":"Ob","name_alt":null,"featureclass":"River"},"geometry":{"type":"LineString","coordinates":[[90.32537072139951,47.650166734398894],[90.20052046098235,47.40801097267918],[89.7334171894961,47.162496242860485],[89.39690026241144,47.06914276792523],[89.04813602087359,47.08743622495655],[86.94356163935473,47.660114447615086],[86.40013227737339,47.85656240497265],[85.2670748229418,47.99438344989507],[84.35684614455272,47.73212555601381],[83.8356376484569,48.0355178899994],[83.61973351434145,48.193983669622426],[83.55115888869014,48.313123684271346],[83.4915759622576,48.778573309924255],[83.61901004428935,48.8964989284143],[83.86840050652995,49.025354112334085],[84.01164757684302,49.17591339781646],[83.88276655470708,49.34623891865047],[83.3637284687675,49.643843492219375],[82.48481570834733,50.06040721292416],[82.14003055209594,50.18215688740382],[81.78108605339659,50.25186839456556],[80.42778364453702,50.3851935898786],[78.54531456900864,50.767289130244706],[77.24286176952427,51.869030666707445],[76.99884565624205,52.13010000264599],[76.6061047708238,52.692520453494595],[76.32787885936435,52.90178416606054],[75.58952599477806,53.29648875590587],[75.30933637889154,53.496321519578544],[75.11782352081784,53.687886054084515],[75.02191206248412,53.8479021266763],[74.99359337758818,53.98559398051799]]}},{"type":"Feature","properties":{"scalerank":2,"name":"Ob","name_alt":null,"featureclass":"River"},"geometry":{"type":"LineString","coordinates":[[74.99359337758818,53.98559398051799],[74.51992719920088,54.375285956431014],[74.39617506296992,54.42505036072808],[74.12368523549486,54.534526882538415],[73.85061404815653,54.64400340434875],[73.72509199411957,54.69376780864583],[73.63321129750463,54.74134888367857],[73.46589592885161,54.82971558289768],[73.37612104685255,54.877296657930415],[73.08177208851805,55.229549058647976],[73.11515506377859,55.30603017843994],[73.33405643096697,55.445763251357164],[73.65517378122865,55.58549632427439],[74.24108117055391,55.76124787049906],[74.47186811716944,55.87488434510888],[74.62452029815964,56.02727814393762],[74.71030317576412,56.203339748756036],[74.70265506378493,56.41234507916052],[74.6298429706857,56.644346421934884],[74.47806928904447,56.859475409565945],[74.27994184763742,57.02778554954058],[73.69021040232255,57.255471910218574],[72.90105960479335,57.446312974672466],[71.47432498564703,57.65875478782435],[70.28902265816834,57.94623078066405],[68.8398604674073,58.03358978945346],[68.68060662219185,58.06849721946662],[68.38450066515941,58.13310567893953],[68.22403242378508,58.16727671979254],[68.24538770942968,58.24941640892047],[68.26811242053003,58.332382920965074],[68.5037052753487,58.52707387962471],[68.60142540881262,58.74607859967766],[68.79314497261547,58.98407440859789],[68.96832807808491,59.40474640566988],[69.12955854683557,59.55042226830068],[69.33430057157597,59.64953766543648],[69.78274865100485,59.87262482364051],[69.804556105432,59.996028143953495],[69.7676591327756,60.10713247338103],[69.83731896350506,60.25686493594672],[69.8040910175414,60.51501455346079],[69.7250777525671,60.65833913842229],[69.49749474475371,60.81308421492031],[69.39806928902414,60.86172465681503],[69.17914208361961,60.96873362898343],[68.95981438586483,61.07570384382761],[68.85913577665218,61.12422801374967],[68.82178663521324,61.106709703202725],[68.78611697782378,61.08934642195267],[68.40846561063475,61.20691030541667],[68.22589277534757,61.44767080346452],[67.64510135281921,61.730470079181785],[67.23282677599465,62.06812388777689],[66.28074018744917,62.44683462182789],[65.68873497911346,62.66237702091729],[65.40994062689882,62.84727529565761],[65.3032287942161,63.0750650092002],[65.42725223171658,63.3065237494355],[65.68899336127492,63.58733348250952],[65.83332563666613,64.01516266545396],[65.91838504421855,64.11505320907413],[65.92603315619775,64.22434886337145],[65.5822815285922,64.7073684760036],[65.51065799343567,64.96564728459838],[65.41102583197693,65.04016469996326],[65.36839277533613,65.16746959091428],[65.76630130398357,65.802676296646],[65.8353926939578,65.99979604762335],[65.8392684263797,66.15717662216825],[66.34140831895982,66.44317983668763],[66.677305128857,66.56653148056832],[67.03309736518653,66.60508209905807],[68.19287153511306,66.57575572373243],[69.013493279908,66.78835256618119],[69.93043989449501,66.74683055283468],[70.84010013212887,66.65303782822492],[71.46957075387618,66.66347646754788]]}},{"type":"Feature","properties":{"scalerank":2,"name":"Peace","name_alt":null,"featureclass":"River"},"geometry":{"type":"LineString","coordinates":[[-124.83563045947423,56.75692352968272],[-124.20045039940291,56.243492336646824],[-123.88873815981833,56.088824774797246],[-123.54353959210863,56.04471893983613],[-122.7300750331861,56.09853994406811],[-122.2217081305148,56.021232001359465],[-121.97363541729766,56.019785061255305],[-121.3571355800556,56.18974884706327],[-120.24459366924387,56.151973375057906],[-119.50944474345967,56.25145050721977],[-119.18871415899591,56.24437083599578],[-118.67034786667612,56.01621938742716],[-118.34623328334149,55.98131195741399],[-117.77913611537048,56.054098212297106],[-117.50612952357251,56.15535818137303],[-117.20036007370149,56.41231924094437],[-117.17689897344098,56.545024319069896],[-117.25821183965225,56.854876207091976],[-117.13227637415694,57.28102590598691],[-117.10752336308914,57.65415558535038],[-117.03701087122688,57.929229234440015],[-116.85415381556209,58.06958242454475],[-116.67623185918117,58.12740835227936],[-116.5421056791676,58.288948879623746],[-116.32315263554695,58.37095937767096],[-115.98847022180863,58.418708401108645],[-114.97251156295039,58.40971670188986],[-114.61323116744114,58.475862535223456],[-114.48895903711025,58.5121652289085],[-114.21144044681779,58.59048086204683],[-113.92362209761417,58.66495952008749],[-113.76845069054974,58.68975128847953],[-113.58348782026906,58.72357351341455],[-113.29752982262794,58.807844855374526],[-113.14918616418008,58.85532257754268],[-112.64761471235515,59.097607530343126],[-112.32267330610385,59.12590037702293],[-112.06504045291271,59.08357737897589],[-111.92647009972205,58.97345490176191],[-111.41076514366532,59.00505504010839],[-111.44383806033211,59.53662466087874],[-111.56902421755917,59.84794932722113],[-112.38313473188535,60.221544094475234],[-112.49320553266703,60.314613349032896],[-112.48369706912533,60.45052236596052],[-112.57134029829234,60.54059438744525],[-113.16011572960981,60.8532368028111],[-113.28245968306082,61.17381155053373],[-113.45965816938966,61.242670396562644],[-113.58197951440154,61.24420777042333],[-113.70401017948177,61.245874335364746],[-113.83890020176014,61.25600291609395],[-114.13603484277289,61.278107510006805],[-114.43324215376855,61.3001862657035],[-114.56835018599566,61.31021149356813],[-116.33434058313813,61.08939809838495],[-116.6986335925797,61.12138580997362],[-117.66810930059138,61.30783437768271],[-118.15497880742927,61.42147085229253],[-119.07063351120897,61.30132314721392],[-119.7683695074581,61.33413768171927],[-120.31781917380145,61.4633546006651],[-120.57503861553425,61.569678860105626],[-120.81339615948052,61.780957953530944],[-121.58916276104611,61.973685207763495],[-123.02421728579144,62.24229930281665],[-123.25466833559705,62.34890778263478],[-123.29611283429513,62.50993154565626],[-123.2289334723157,62.86667979598322],[-123.3205041103369,63.05733999292407],[-123.9544189052613,63.74101919214557],[-124.33648860741127,64.00276032170389],[-124.48934749413064,64.22316030542875],[-124.7616822923088,64.42387156845037],[-125.02515458234892,64.69202057561291],[-125.30022823143857,64.80596710881649],[-126.18162146060875,65.04910472274976],[-126.92932775943987,65.2913896755502],[-128.03019079655365,65.60834707301241],[-128.67141780664736,65.72275869410662],[-128.90856095443476,65.86055390081289],[-128.98336259017725,66.02966502548806],[-128.84207922429127,66.3225153672861],[-129.02560807357585,66.45547882757309],[-129.77052384506322,66.7316118435247],[-130.0437113043742,66.88087921819978],[-130.34503658106811,67.21377879502403],[-130.5328287360168,67.30584035915201],[-130.75940405340052,67.35366689723814],[-131.20867895574605,67.45469432236875],[-132.65282853257474,67.29330882432124],[-133.01577633988822,67.29274038356603],[-133.3725504284313,67.35935130479025],[-133.8297318249175,67.52058177354088],[-134.19007158728874,67.74152435980479],[-134.39956784379999,68.05264232041803],[-134.40719011756306,68.18209178330916],[-134.15968584510114,68.4496723497165],[-134.24422848833063,68.70619415961333],[-134.83589779985644,68.9637753363722],[-135.3134138724495,69.37499054633476]]}}]}

	glify.lines({
		map: map,
		size: 4,
		data: data,

		click: (e: L.LeafletMouseEvent, feature: any) => {
			//set up a standalone popup (use a popup as a layer)
			console.log('clicked on Point', feature, e);

			L.popup()
				.setLatLng(e.latlng)
				.setContent(`You clicked the point at longitude:${ e.latlng.lng }, latitude:${ e.latlng.lat }`)
				.openOn(map);
		
		},

		color: (index: number) => {
			const r = {
				r: Math.random(),
				g:  (7 * index) / 255.0,
				b:  (index * 9) / 255.0,
				a: 1,
			};

			return r
		},

		hover: (e: L.LeafletMouseEvent, feature: any) => {
			console.log('hovered on Line ', feature, e);
		},
		hoverOff: (e: L.LeafletMouseEvent, feature: any) => {
			console.log('hovered off Line', feature, e);
		},
		hoverWait: 100
	})
}


/*
* Create universal marker with a contextmenu to remove 
*/
function createMarker(map: L.Map, position: L.LatLng, options?: L.MarkerOptions) {

	// marker.svg
	//var svg = 'b2xvcj0iIzJlNmM5NyIgb2Zmc2V0PSIwIi8+++++CiA8L2c+Cjwvc3ZnPg==+CiAgPGxpbmVhckdyYWRpZW50IGlkPSJiIj4KICAgPHN0b3Agc3RvcC1jb2xvcj0iIzJlNmM5NyIgb2Zmc2V0PSIwIi8+CiAgIDxzdG9wIHN0b3AtY29sb3I9IiMzODgzYjciIG9mZnNldD0iMSIvPgogIDwvbGluZWFyR3JhZGllbnQ+CiAgPGxpbmVhckdyYWRpZW50IGlkPSJhIj4KICAgPHN0b3Agc3RvcC1jb2xvcj0iIzEyNmZjNiIgb2Zmc2V0PSIwIi8+CiAgIDxzdG9wIHN0b3AtY29sb3I9IiM0YzljZDEiIG9mZnNldD0iMSIvPgogIDwvbGluZWFyR3JhZGllbnQ+CiAgPGxpbmVhckdyYWRpZW50IHkyPSItMC4wMDQ2NTEiIHgyPSIwLjQ5ODEyNSIgeTE9IjAuOTcxNDk0IiB4MT0iMC40OTgxMjUiIGlkPSJjIiB4bGluazpocmVmPSIjYSIvPgogIDxsaW5lYXJHcmFkaWVudCB5Mj0iLTAuMDA0NjUxIiB4Mj0iMC40MTU5MTciIHkxPSIwLjQ5MDQzNyIgeDE9IjAuNDE1OTE3IiBpZD0iZCIgeGxpbms6aHJlZj0iI2IiLz4KIDwvZGVmcz4KIDxnPgogIDx0aXRsZT5MYXllciAxPC90aXRsZT4KICA8cmVjdCBpZD0ic3ZnXzEiIGZpbGw9IiNmZmYiIHdpZHRoPSIxMi42MjUiIGhlaWdodD0iMTQuNSIgeD0iNDExLjI3OSIgeT0iNTA4LjU3NSIvPgogIDxwYXRoIHN0cm9rZT0idXJsKCNkKSIgaWQ9InN2Z18yIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS13aWR0aD0iMS4xIiBmaWxsPSJ1cmwoI2MpIiBkPSJtMTQuMDk1ODMzLDEuNTVjLTYuODQ2ODc1LDAgLTEyLjU0NTgzMyw1LjY5MSAtMTIuNTQ1ODMzLDExLjg2NmMwLDIuNzc4IDEuNjI5MTY3LDYuMzA4IDIuODA2MjUsOC43NDZsOS42OTM3NSwxNy44NzJsOS42NDc5MTYsLTE3Ljg3MmMxLjE3NzA4MywtMi40MzggMi44NTIwODMsLTUuNzkxIDIuODUyMDgzLC04Ljc0NmMwLC02LjE3NSAtNS42MDcyOTEsLTExLjg2NiAtMTIuNDU0MTY2LC0xMS44NjZ6bTAsNy4xNTVjMi42OTE2NjcsMC4wMTcgNC44NzM5NTgsMi4xMjIgNC44NzM5NTgsNC43MXMtMi4xODIyOTIsNC42NjMgLTQuODczOTU4LDQuNjc5Yy0yLjY5MTY2NywtMC4wMTcgLTQuODczOTU4LC0yLjA5IC00Ljg3Mzk1OCwtNC42NzljMCwtMi41ODggMi4xODIyOTIsLTQuNjkzIDQuODczOTU4LC00LjcxeiIvPgogIDxwYXRoIGlkPSJzdmdfMyIgZmlsbD0ibm9uZSIgc3Ryb2tlLW9wYWNpdHk9IjAuMTIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS13aWR0aD0iMS4xIiBzdHJva2U9IiNmZmYiIGQ9Im0zNDcuNDg4MDA3LDQ1My43MTljLTUuOTQ0LDAgLTEwLjkzOCw1LjIxOSAtMTAuOTM4LDEwLjc1YzAsMi4zNTkgMS40NDMsNS44MzIgMi41NjMsOC4yNWwwLjAzMSwwLjAzMWw4LjMxMywxNS45NjlsOC4yNSwtMTUuOTY5bDAuMDMxLC0wLjAzMWMxLjEzNSwtMi40NDggMi42MjUsLTUuNzA2IDIuNjI1LC04LjI1YzAsLTUuNTM4IC00LjkzMSwtMTAuNzUgLTEwLjg3NSwtMTAuNzV6bTAsNC45NjljMy4xNjgsMC4wMjEgNS43ODEsMi42MDEgNS43ODEsNS43ODFjMCwzLjE4IC0yLjYxMyw1Ljc2MSAtNS43ODEsNS43ODFjLTMuMTY4LC0wLjAyIC01Ljc1LC0yLjYxIC01Ljc1LC01Ljc4MWMwLC0zLjE3MiAyLjU4MiwtNS43NjEgNS43NSwtNS43ODF6Ii8+CiA8L2c+Cjwvc3ZnPg=='; /* insert your own svg */
	//var iconUrl = 'data:image/svg+xml;base64,' + svg;

	const marker = new L.Marker(position, {
		title: options?.title ?? "",
		icon: L.icon({
			iconUrl: '/static/img/svg/marker.svg',
			iconSize: [28, 41]
		}),

		contextmenu: true,
		contextmenuItems: [
			{
				text: 'Remove marker',
				callback: () => {
					map.removeLayer(marker);
				}
			},
			{
				text: 'Unlock position',
				callback: () => {
					marker?.dragging?.enable();
					// TODO: lock again / modify contextmenu
				}
			},
			{
				text: "",
				separator: true
			},
			{
				text: 'Show coordinates',
				callback: function () {
					showCoordinates(position)
				}
			},
		]
	});

	return marker;
}

/*
* Displays the given coordinates in an alter window
*/
function showCoordinates(position: L.LatLng) {
	alert([(0|position.lat*1000000)/1000000, ' ', (0|position.lng*1000000)/1000000, "\n\n", convertDDtoDM(position.lat, position.lng)].join(''));
}

/*
* Convert decimal degree to decimal minutes
*/
function convertDDtoDM(lat: number, lon: number) {
	function helper(x: number, lon: boolean) {
		return [
			x<0?lon?'W':'S':lon?'E':'N',
			' ',
			0|Math.abs(x),
			'° ',
			(0|Math.abs(x)%1*60000)/1000
		].join('');
	}

	return helper(lat, false) + ' ' + helper(lon, true);
}