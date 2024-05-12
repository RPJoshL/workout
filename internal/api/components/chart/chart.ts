import echarts, { EChartsOption, TooltipComponentFormatterCallbackParams } from "echarts";
import L from "leaflet";


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

interface WDetails {
	TitleSpeed: string
	TitleElevation: string
	TitleHeartrate: string

	Points: WDetailsPoint[]
}

interface WDetailsPoint {
	Duration: number

	Speed: number
	Elevation: number
	HeartRate: number

	Latitude: number
	Longitude: number
}

/** Returns the line series data for the heart rate diagramm  */
function getHeartrateLines(data: WDetails): echarts.SeriesOption[] {

	// Base options to apply to every instance
	const baseOptions: echarts.SeriesOption = {
		name: 'heartrate',
		yAxisIndex: 2,
		xAxisIndex: 2,
		type: 'line',
		smooth: true,
		connectNulls: true,
		showSymbol: false,
	}

	/** Duration offset in seconds to use for many datapoints on mobile devices */
	let offsetDuration = (window.devicePixelRatio > 1 && data.Points.length > 600) ? ( (data.Points.length - 600) / 10) : 0
	if (offsetDuration > 120) offsetDuration = 120;
	if (offsetDuration > 0) console.log("Using a color downsampling duration of " + offsetDuration + " seconds (" + window.devicePixelRatio + " ; " + data.Points.length + ")")

	/** Ignore color change checks if a color change should be ignored if the last color is only "temporary" */
	const ignoreColorChange = (startIndex: number, currentColor: string, pt: WDetailsPoint): boolean => {
		const maxAllowedOffset = 36 + offsetDuration
		let otherColorFound = false
		for (let ii = startIndex; ii < data.Points.length; ii++) {
			const duration = data.Points[ii].Duration - pt.Duration
			if (duration > maxAllowedOffset * 2) {
				break
			}
				
			// Same color found again
			const nextColor = GetColorByHeartrate(data.Points[ii].HeartRate)
			if (nextColor == currentColor && (duration < maxAllowedOffset || ii < startIndex + 1)) {
				otherColorFound = true
			}
		}
	
		return otherColorFound
	}

	// Return data
	const rtc: echarts.SeriesOption[] = []

	// Current data that are processed
	let currentData: any[] = []
	/** Index to which color comparisons are ignored (excluding) */
	let ignoreIndex = 0
	let prevPoint: WDetailsPoint | null = null
	let lastColor = GetColorByHeartrate(data.Points[0].HeartRate)
	data.Points.forEach( (p, i) => {
		if (prevPoint) {
			const isLastPoint = i == data.Points.length - 1

			// Render the last line of a color level
			const currentColor = GetColorByHeartrate(p.HeartRate)
			if ( (lastColor != currentColor && i >= ignoreIndex) || isLastPoint) {
				
				// The color is different. But we don't change the color for points that are
				// only present for distances < 200 meters and changes back to the default color
				// again!
				const otherColorFound = ignoreColorChange(i, lastColor, p)
				if (otherColorFound && !isLastPoint) {
					currentData.push([p.Duration, p.HeartRate])
				} else {

					// Render last points
					currentData.push([ p.Duration, p.HeartRate ])
					// Also add the next point to fill spaces procuded by downsampling.
					// If not downsampling this line will be overlayed by the "correct" color
					if (i + 2 < data.Points.length && offsetDuration > 20) {
						const nextPoint = data.Points[i+1]
						currentData.push( [ nextPoint.Duration, nextPoint.HeartRate ] )
					}

					rtc.push({
						...baseOptions,
						data: currentData.map((val: any) => [ val[0] / 60.0, val[1] ]),
						itemStyle: {
							color: lastColor,
						},
						emphasis: {
							itemStyle: {
								color: lastColor
							},
							areaStyle: {
								color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
									{
										offset: 0,
										color: lastColor + "cc" // .8
									},
									{
										offset: 1,
										color: lastColor + "80" // .3
									}
								])
							}
						},
						areaStyle: {
							color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
								{
									offset: 0,
									color: lastColor + "cc" // .8
								},
								{
									offset: 1,
									color: lastColor + "80" // .3
								}
							])
						},
					})

					// Reassigne data
					currentData = [ [ p.Duration, p.HeartRate ] ]
					lastColor = currentColor

					// If the new heart rate zone is only "valid" for a few meters and followed
					// by a "higher" zone, add the "higher" zone directly instead of displaying the
					// old one
					let zoneSmall = false
					let sameColor = true
					let newColor: string = ""
					let newIgnoreIndex = 0
					for (let ii = i+1; ii < data.Points.length; ii++) {
						const duration = data.Points[ii].Duration - p.Duration 

						const nextColor = GetColorByHeartrate(data.Points[ii].HeartRate)
						if (!zoneSmall && nextColor != currentColor && duration < (30 + offsetDuration)) {
							zoneSmall = true
							newColor = nextColor
							newIgnoreIndex = ii
						}

						if (zoneSmall && newColor != nextColor && !ignoreColorChange(ii, newColor, data.Points[ii]) && (duration + offsetDuration) < 42) {
							sameColor = false
						}

						if ( (duration + offsetDuration) > 66) {
							break
						}
					}
					
					if (zoneSmall && sameColor) {
						lastColor = newColor
						ignoreIndex = newIgnoreIndex
					}
				}
			} else {
				currentData.push([p.Duration, p.HeartRate])
			}
		} else {
			currentData.push([p.Duration, p.HeartRate])
		}

		prevPoint = p
	}) 

	return rtc
}

export function AddWorkoutDetailsChart(id: string, darkTheme: boolean, data: WDetails) {
	// @ts-expect-error Declared globally
	OnElementReady("#" + id, () => {
		addWorkoutDetailsChart(id, darkTheme, data)
	})
}

/**
 * Initialize and configure a new apache Echart
 */
export function addWorkoutDetailsChart(id: string, darkTheme: boolean, data: WDetails) {

	// Specify the configuration items and data for the chart
	const xCutOffSeconds = data.Points[data.Points.length - 1].Duration > 5500
	
	/** Format the x Axis to show minutes */
	const formatDurationAxis = (value: number) => {
		const duration = value * 60
	
		// Only seconds
		if (duration < 60) {
			if (duration === 0) return "0s"
			return String(duration).padStart(2, '0') + "s"
		}

		// Show minutes
		if ( duration < 3600 ) {
			let m = (duration / 60).toFixed(0) + "m"
			if (duration % 60 !== 0 && !xCutOffSeconds) m += " " + (duration % 60).toFixed(0) + "s" 
			return m
		}

		// Show hours with minutes
		return Math.trunc(duration / 3600) + "h " + String( ((duration / 60) % 60).toFixed(0) ).padStart(2, '0') + "m"
	}

	/** Weather to ignore elevation */
	const ignoreElevation = data.Points.filter(e => e.Elevation < 1).length > (data.Points.length / 2)
	const ignoreHeartrate = data.Points.filter(e => e.HeartRate < 20).length > (data.Points.length / 2)
	const offsetElevation = ignoreElevation ? (ignoreHeartrate ? 26 : 13) : 0
	const offsetHeartRate = ignoreHeartrate ? (ignoreElevation ? 34 : 17) : 0
	const imgExtension = darkTheme ? "-dark.svg" : ".svg"

	// Get an initialized leaflet map
	const map: L.Map | undefined = (window as any)[Object.keys(window as any).filter(key => key.substr(0,11) === "leaflet-map").at(-1) ?? ""]
	let lastMarker: L.CircleMarker | null = null

	const option: EChartsOption = {
		animationDuration: 250,

		title: [
			{
				top: '2%',
				left: window.innerWidth < 600 ? '6px' : '12px',

				text: "{Speed|}  " + data.TitleSpeed,
				textStyle: {
					rich: {
						Speed: {
							height: 24,
							align: 'left',
							backgroundColor: {
								image: '/static/img/svg/speed' + imgExtension
							},
						}
					},
					fontSize: "1rem"
				}
			},
			{
				top: 36 + offsetHeartRate + '%',
				left: window.innerWidth < 600 ? '6px' : '12px',

				text: "{Elevation|}  " + data.TitleElevation,
				textStyle: {
					rich: {
						Elevation: {
							height: 25,
							align: 'left',
							backgroundColor: {
								image: '/static/img/svg/mountain' + imgExtension
							}
						}
					},
					fontSize: "1rem",
				},
				show: !ignoreElevation,
			},
			{
				top: 63 - offsetElevation + '%',
				left: window.innerWidth < 600 ? '6px' : '12px',

				text: "{Heartrate|}  " + data.TitleHeartrate,
				textStyle: {
					rich: {
						Heartrate: {
							height: 26,
							align: 'left',
							backgroundColor: {
								image: '/static/img/svg/heartrate' + imgExtension
							}
						}
					},
					fontSize: "1rem",
				},
				show: !ignoreHeartrate,
			}
		],


		tooltip: {
			trigger: "axis",
			formatter: (param: TooltipComponentFormatterCallbackParams) => {

				// Point the user is currently hoovering over
				const point = data.Points[param[0].dataIndex]

				// Get leaflet map to synchronize 
				if (map !== undefined) {
					if (lastMarker !== null) map.removeLayer(lastMarker)
					lastMarker = L.circleMarker([point.Latitude, point.Longitude], {
						color: "#0852c7",
						radius: 4,
						fill: true,
						fillColor: '#3080ff',
						fillOpacity: 1,
						pane: "overlay-pane"
					}).addTo(map).bringToFront()
				}

				// Build header duration
				const duration = point.Duration
				let header = ""
				if (duration < 60) header = String(duration).padStart(2, '0') + "s"
				else if ( duration < 3600 ) header = (duration / 60).toFixed(0) + "m " +  String(duration % 60).padStart(2, '0') + "s"
				else header = Math.trunc(duration / 3600) +"h " + String( ((duration / 60) % 60).toFixed(0) ).padStart(2, '0') + "m"

				// Add every graph to result
				let rtc = `<b>${header}<b>`;
				let heartRatePresent = false;
				(param as any).forEach( (p: any) => {
					// Value to display
					const val = (p.value[1] as number) || 0

					// Ignore fields
					const seriesName = p.seriesName
					if (ignoreElevation && seriesName === "elevation") return ""
					if (ignoreHeartrate && seriesName === "heartrate") return ""
					if (seriesName === "heartrate-hide") return ""

					// Get unit name
					let unit = "km/h"
					if (seriesName === "elevation") unit = "m"
					else if (seriesName === "heartrate") unit = "bpm"

					// The heart rate could be shown doubled when downsampling
					if (seriesName === "heartrate") {
						if (heartRatePresent) return ""
						else heartRatePresent = true
					}

					const value = (val % 1 !== 0) ? parseFloat(val as any).toFixed(2) : parseFloat(val as any).toFixed(0)
					rtc += `<br> ${p.marker} ${value} ${unit}`
				})

				return rtc
			},
			axisPointer: {
				animation: false
			},
		},

		grid: [
			{ left: window.innerWidth < 600 ? '36px' : '50px', right: window.innerWidth < 600 ? '24px' : '40px', top: '52px', width: 'auto', height: 23 + offsetElevation + offsetHeartRate + '%' },
			{ left: window.innerWidth < 600 ? '36px' : '50px', right: window.innerWidth < 600 ? '24px' : '40px', top: 43 + offsetHeartRate + '%', width: 'auto', height: ignoreElevation ? '0' : 15 + offsetHeartRate + '%' },
			{ left: window.innerWidth < 600 ? '36px' : '50px', right: window.innerWidth < 600 ? '24px' : '40px', top: 71 - offsetElevation + '%', width: 'auto', height: ignoreHeartrate ? '0' : 23 + offsetElevation + '%' },
		],

		xAxis: [
			{
				type: 'value',
				axisLabel: {
					formatter: formatDurationAxis,
				},
				max: Math.ceil(data.Points[data.Points.length - 1].Duration / 60 ),
				gridIndex: 0,
			},
			{
				show: !ignoreElevation,
				type: 'value',
				axisLabel: {
					formatter: formatDurationAxis,
				},
				max: Math.ceil(data.Points[data.Points.length - 1].Duration / 60 ),
				gridIndex: 1,
			},
			{
				show: !ignoreHeartrate,
				type: 'value',
				axisLabel: {
					formatter: formatDurationAxis,
				},
				max: Math.ceil(data.Points[data.Points.length - 1].Duration / 60 ),
				gridIndex: 2,
			},
		],

		yAxis: [
			{
				type: 'value',
				scale: true,
				gridIndex: 0,
			},
			{
				show: !ignoreElevation,
				type: 'value',
				scale: true,
				gridIndex: 1,
			},
			{
				show: !ignoreHeartrate,
				type: 'value',
				scale: true,
				gridIndex: 2,
			},
		],

		// Synchronize all series
		axisPointer: {
			link: [
				{
					xAxisIndex: 'all'
				}
			]
		},

		series: ([
			{
				name: 'speed',
				type: 'line',
				data: data.Points.map(p => [ p.Duration / 60.0, p.Speed <= 0 ? null : p.Speed ]),
				//connectNulls: true,
				/*
				markPoint: {
					data: [
						{ 
							type: 'max', name: 'Max', label: {
								formatter: (param) => {
									return (param.value as number).toFixed(0)
								}
							}
						},
					]
				},
				*/
				markLine: {
					data: [{ type: 'average', name: 'Avg', label: { 
						formatter: (val) => { 
							return (val.value as number).toFixed(0) 
						},
					}}]
				},

				sampling: "average",
				smooth: 10,
				connectNulls: true,
				showSymbol: false,

				areaStyle: {
					color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
						{
							offset: 0,
							color: 'rgba(58,77,233,0.8)'
						},
						{
							offset: 1,
							color: 'rgba(58,77,233,0.3)'
						}
					])
				},
			},
			{
				name: 'elevation',
				yAxisIndex: 1,
				xAxisIndex: 1,
				type: 'line',
				data: data.Points.map(p => [ p.Duration / 60.0, p.Elevation <= 0 ? null : p.Elevation ]),
				sampling: "average",
				smooth: true,
				connectNulls: true,
				showSymbol: false,
				areaStyle: {
					color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
						{
							offset: 0,
							color: 'rgba(69,122,2,0.8)'
						},
						{
							offset: 1,
							color: 'rgba(69,122,2,0.3)'
						}
					])
				},
				markLine: {
					data: [{ type: 'average', name: 'Avg', label: { 
						formatter: (val) => { 
							return (val.value as number).toFixed(0) 
						},
					}}]
				},

				itemStyle: {
					color: '#59a102'
				},
			},
			{
				name: 'heartrate-hide',
				yAxisIndex: 2,
				xAxisIndex: 2,
				type: 'custom',
				data: data.Points.map(p => [ p.Duration / 60.0, p.HeartRate <= 0 ? 0 : p.HeartRate, p.HeartRate ?? 20 ]),
				markLine: {
					data: [{ type: 'average', name: 'Avg', label: { 
						formatter: (val) => { 
							return (val.value as number).toFixed(0) 
						},
					}}]
				},
			},
		] as echarts.SeriesOption[]).concat(getHeartrateLines(data) as any) as any,

		// Support zooming without displaying toolbox: https://github.com/apache/echarts/issues/13397
		toolbox : {
			orient   : 'horizontal',
			left: 'center',
			itemSize : 13,
			top      : 22,
			right    : -6,
			feature  : {
				dataZoom: {
					yAxisIndex: "none",
					xAxisIndex: [0,1,2],
					icon : {
						// zoom : 'path://', // hack to remove zoom button
						// back : 'path://', // hack to remove restore button
					},
				},
				restore: {}
			},
		},

		dataZoom: [
			{
				type: 'inside',
				throttle: 50,
				xAxisIndex: [0,1,2],
			}
		],
	}
	
	// Display chart
	renderChart(id, darkTheme, option).dispatchAction({
		type                    : 'takeGlobalCursor',
		key                     : 'dataZoomSelect',
		dataZoomSelectActive    : false,
	});
}

function renderChart(id: string, darkTheme: boolean, options: EChartsOption): echarts.ECharts {
	
	// Initialize chart
	const divElement = document.getElementById(id)
	const chart = echarts.init(divElement, darkTheme ? "dark" : undefined, {
		renderer: 'canvas'
	})

	// Display the chart using the configuration items and data just specified.
	chart.setOption(options);

	// Redraw chart when size changes
	window.addEventListener("resize", () => {
		setTimeout(() => chart.resize(), 300)
	})

	let initialFinish = true
	chart.on('finished', () => {
		if (!initialFinish) return
		initialFinish = false

		// For chrome: images in rich text are not loaded if cached!
		chart.resize()
	})

	return chart
}