import L from "leaflet";

export class BingLayer extends L.TileLayer{
	getTileUrl(coords) {
		const quadkey = this.toQuadKey(coords.x, coords.y, coords.z)
		let url = L.Util.template((this as any)._url, {
			q: quadkey,
			s: (this as any)._getSubdomain(coords)
		})
		if (typeof (this.options as any).style === 'string') {
			url += '&st=' + (this.options as any).style
		}
		return url
	}

	toQuadKey(x, y, z) {
		let index = ''
		for (let i = z; i > 0; i--) {
			let b = 0
			const mask = 1 << (i - 1)
			if ((x & mask) !== 0) b++
			if ((y & mask) !== 0) b += 2
			index += b.toString()
		}
		return index
	}
}

/**
 * SmoothPoly renders a polyline with rounded corners
 */
export class SmoothPoly extends L.Polyline{

	// override default method to use custom points-to-path method 
	_updatePath() {
		const path = this.roundPathCorners( (this as any)._parts, -.15);
		(this as any)._renderer._setPath(this, path);
	}

	roundPathCorners(rings, radius) {
		function moveTowardsFractional(movingPoint, targetPoint, fraction) {
			return {
				x: movingPoint.x + (targetPoint.x - movingPoint.x) * fraction,
				y: movingPoint.y + (targetPoint.y - movingPoint.y) * fraction
			};
		}
			
		function pointForCommand(cmd) {
			return {
				x: parseFloat(cmd[cmd.length - 2]),
				y: parseFloat(cmd[cmd.length - 1])
			};
		}
			
		const resultCommands: Array<any> = [];
		if (+radius) {
			// negative numbers create artifacts
			radius = Math.abs(radius);
		} else {
			radius = 0.15;
		}
			
		let commands: Array<any> = []
		for (let i = 0, len = rings.length; i < len; i++) {
			commands = rings[i];
			// start point    
			resultCommands.push(["M", commands[0].x, commands[0].y]);
				
			for (let cmdIndex = 1; cmdIndex < commands.length; cmdIndex++) {
				const prevCmd = resultCommands[resultCommands.length - 1];
				let curCmd = commands[cmdIndex];
				const nextCmd = commands[cmdIndex + 1];
			
				if (nextCmd && prevCmd) {
					// Calc the points we're dealing with
					const prevPoint = pointForCommand(prevCmd); // convert to Object
					const curPoint = curCmd;
					const nextPoint = nextCmd;
					
					// The start and end of the cuve are just our point moved towards the previous and next points, respectivly						
					const curveStart = moveTowardsFractional(
						curPoint,
						prevCmd.origPoint || prevPoint,
						radius
					);
					const curveEnd = moveTowardsFractional(
						curPoint,
						nextCmd.origPoint || nextPoint,
						radius
					);
					
					// Adjust the current command and add it
					curCmd = Object.values(curveStart);
					
					curCmd.origPoint = curPoint;
					curCmd.unshift("L");
					resultCommands.push(curCmd);
					
					// The curve control points are halfway between the start/end of the curve and
					// calculate curve, if radius is different than 0
					if (radius) {
						const startControl = moveTowardsFractional(curveStart, curPoint, 0.5);
						const endControl = moveTowardsFractional(curPoint, curveEnd, 0.5);
						// Create the curve
						const curveCmd: any = [
							"C",
							startControl.x,
							startControl.y,
							endControl.x,
							endControl.y,
							curveEnd.x,
							curveEnd.y
						];
						// Save the original point for fractional calculations
						curveCmd.origPoint = curPoint;
						resultCommands.push(curveCmd);
					}
				} else {
					// Pass through commands that don't qualify
					const el = Object.values(curCmd);
					el.unshift("L");
					resultCommands.push(el);
				}
			}
		}
			
		return (
			resultCommands.reduce(function(str, c) {
				return str + c.join(" ") + " ";
			}, "") || "M0 0"
		);
	}
}

/*
 * https://github.com/adoroszlai/leaflet-distance-markers
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2014- Doroszlai Attila, 2016- Phil Whitehurst
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

export class DistanceMarkers extends L.LayerGroup {

	constructor (line: L.Polyline, map:L.Map, options: L.DistanceOptions) {
		super(line as any, undefined)
		options = options || {};
		const offset = options.offset || 1000;
		const showAll = Math.min(map.getMaxZoom(), options.showAll || 12);
		const cssClass = options.cssClass || 'dist-marker';
		const iconSize = options.iconSize !== undefined ? options.iconSize : L.point(12, 12);
		const textFunction = options.textFunction || function(distance, i) {
			return i;
		};

		const zoomLayers = {};
		// Get line coords as an array
		let coords = line;
		if (typeof line.getLatLngs == 'function') {
			coords = (line as any).getLatLngs();
		}
		// Get accumulated line lengths as well as overall length
		const accumulated = L.GeometryUtil.accumulatedLengths(line);
		const length = accumulated.length > 0 ? accumulated[accumulated.length - 1] : 0;
		// Position in accumulated line length array
		let j = 0;
		// Number of distance markers to be added
		const count = Math.floor(length / offset);

		for (let i = 1; i <= count; ++i) {
			const distance = offset * i;
			// Find the first accumulated distance that is greater
			// than the distance of this marker
			while (j < accumulated.length - 1 && accumulated[j] < distance) {
				++j;
			}
			// Now grab the two nearest points either side of
			// distance marker position and create a simple line to
			// interpolate on
			const p1 = coords[j - 1];
			const p2 = coords[j];
			const m_line = L.polyline([p1, p2]);
			const ratio = (distance - accumulated[j - 1]) / (accumulated[j] - accumulated[j - 1]);
			const position = L.GeometryUtil.interpolateOnLine(map, m_line, ratio);
			const text = textFunction.call(this, distance, i, offset);
			const icon = L.divIcon({ className: cssClass, html: text, iconSize: iconSize });
			const marker = L.marker(position!.latLng, { title: text, icon: icon, contextmenu: false, contextmenuItems: [] });

			// visible only starting at a specific zoom level
			const zoom = this._minimumZoomLevelForItem(i, showAll);
			if (zoomLayers[zoom] === undefined) {
				zoomLayers[zoom] = L.layerGroup();
			}
			zoomLayers[zoom].addLayer(marker);
		}

		let currentZoomLevel = 0;

		const updateMarkerVisibility = () => {
			const oldZoom = currentZoomLevel;
			const newZoom = currentZoomLevel = map.getZoom();

			if (newZoom > oldZoom) {
				for (let i = oldZoom + 1; i <= newZoom; ++i) {
					if (zoomLayers[i] !== undefined) {
						this.addLayer(zoomLayers[i]);
					}
				}
			} else if (newZoom < oldZoom) {
				for (let i = oldZoom; i > newZoom; --i) {
					if (zoomLayers[i] !== undefined) {
						this.removeLayer(zoomLayers[i]);
					}
				}
			}
		};
		map.on('zoomend', updateMarkerVisibility);

		(this as any)._layers = {}; // need to initialize before adding markers to this LayerGroup
		updateMarkerVisibility();
	}

	_minimumZoomLevelForItem(item, showAllLevel) {
		let zoom = showAllLevel;
		let i = item;
		while (i > 0 && i % 2 === 0) {
			--zoom;
			i = Math.floor(i / 2);
		}
		return zoom;
	}

}

L.Polyline.include({

	_originalOnAdd: L.Polyline.prototype.onAdd,
	_originalOnRemove: L.Polyline.prototype.onRemove,
	_originalSetLatLngs: L.Polyline.prototype.setLatLngs,

	addDistanceMarkers: function () {
		if (this._map && this._distanceMarkers) {
			this._map.addLayer(this._distanceMarkers);
		}
	},

	removeDistanceMarkers: function () {
		if (this._map && this._distanceMarkers) {
			this._map.removeLayer(this._distanceMarkers);
		}
	},

	onAdd: function (map) {
		this._originalOnAdd(map);
		this.createDistanceMarkers();
	},

	createDistanceMarkers: function() {
		if (!this._map) return;

		const opts: L.DistanceOptions | undefined = this.options.distanceMarkers;
		
		// Only add markers if object is defined
		if (opts === undefined || opts === null) return

		if (this._distanceMarkers === undefined && this.options.distanceMarkers) {
			this._distanceMarkers = new DistanceMarkers(this, this._map, opts);
		}
		if (opts.lazy === undefined || opts.lazy === false) {
			this.addDistanceMarkers();
		}
	},

	destroyDistanceMarkers: function() {
		if (this._distanceMarkers) {
			this._distanceMarkers = undefined;
		}
	},

	setLatLngs: function (latlngs) {
		const recreate = this._map && this._distanceMarkers;
		this.removeDistanceMarkers();
		this.destroyDistanceMarkers();
		const result = this._originalSetLatLngs(latlngs);
		if (recreate) {
			this.createDistanceMarkers();
		}
		return result;
	},

	onRemove: function (map) {
		this.removeDistanceMarkers();
		this._originalOnRemove(map);
	}

});

declare module 'leaflet' {

	export interface DistanceOptions {

		/** Distance in meters between the markers  */
		offset?: number;
	
		/** The zoom level at which all distance markers will be shown.
		 * Zooming out once from this level will remove approximately half of the markers (default: 12)
		 */
		showAll?: number;
	
		/** Don't create markers by default. Markers are only added if Polyline.addDistanceMarkers is explicitly called */
		lazy?: boolean;
	
		/** Size of the marker icon in pixels. Set to null to allow sizing via CSS  */
		iconSize?: L.PointExpression
	
		textFunction?: (distance: number, index: number, offset: number) => void
	
		cssClass?: string
	}

	interface PolylineOptions {
        distanceMarkers?: DistanceOptions
	}
}