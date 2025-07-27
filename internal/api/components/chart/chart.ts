import echarts, { EChartsOption, TooltipComponentFormatterCallbackParams } from "echarts";
import { MarkLine1DDataItemOption } from "echarts/types/src/component/marker/MarkLineModel.js";
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

	Part: number
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

type ElementDetails = {
	percent: number
	hide: boolean
	titleTop?: number
	top?: number
	height?: number
}

type ElementsType = Record<"speed" | "elevation" | "heartrate", ElementDetails>

function getChartElements(data: WDetails): ElementsType {
	const elements: ElementsType = {
		"speed": { percent: 0.23, hide: data.Points.filter(e => e.Speed < 0.5).length > (data.Points.length / 2), height: 0 },
		"elevation": { percent: 0.15, hide:  data.Points.filter(e => e.Elevation < 1).length > (data.Points.length / 2) },
		"heartrate": { percent: 0.23, hide: data.Points.filter(e => e.HeartRate < 20).length > (data.Points.length / 2) },
	}

	// The gap between the title and the element
	const titleGap = 0.08
	// The gap between two elements (with title)
	const elementGap = 0.03

	// Margin at the top of the chart (end = 0.03)
	const startMargin = 0.02

	// Calculate the additional percentages we do have because an element is hidden
	const freePercentage = Object.entries(elements).filter( ([_, val]) => val.hide ).map( ([_, val]) => val.percent + titleGap + (elementGap / 2) ).reduce( (a, b) => a + b, 0 )
	const shownElements = Object.keys(elements).filter( (key) => !(elements as any)[key].hide )

	// Add these percentages to the elements
	let currentOffset = startMargin * 100
	shownElements.forEach( (key) => {
		const el = (elements as any)[key] as ElementDetails
		el.height = el.percent + (freePercentage / shownElements.length)
		el.height = Math.round(el.height * 100)

		el.titleTop = currentOffset
		el.top = Math.round(currentOffset + (titleGap * 100))

		currentOffset = el.top + el.height + (elementGap * 100)
	})

	return elements
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

	const imgExtension = darkTheme ? "-dark.svg" : ".svg"
	const elements = getChartElements(data)

	// Get an initialized leaflet map
	const map: L.Map | undefined = (window as any)[Object.keys(window as any).filter(key => key.substr(0,11) === "leaflet-map").at(-1) ?? ""]
	let lastMarker: L.CircleMarker | null = null

	/** Get vertical lines to indicate different track segments */
	const segmentMarker: (MarkLine1DDataItemOption)[] = []
	let lastPart = data.Points[0].Part
	data.Points.forEach(p => {
		if (p.Part != lastPart) {
			lastPart = p.Part

			segmentMarker.push({
				xAxis: p.Duration / 60.0,
				lineStyle: {
					type: 'dashed',
					width: 2,
					opacity: 0.7
				},
				label: { formatter: (p.Part+1) + "." },
				symbolSize: [0,0],
				itemStyle: {
					opacity: 0.5
				}
			})
		}
	})

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
				},
				show: !elements["speed"].hide,
			},
			{
				top: elements["elevation"].titleTop + '%',
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
				show: !elements["elevation"].hide,
			},
			{
				top: elements["heartrate"].titleTop + '%',
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
				show: !elements["heartrate"].hide,
			}
		],


		tooltip: {
			trigger: "axis",
			formatter: (param: TooltipComponentFormatterCallbackParams) => {

				// Point the user is currently hoovering over
				const point = Array.isArray(param) ? data.Points[param[0].dataIndex] : data.Points[param.dataIndex]

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
					if (elements["elevation"].hide && seriesName === "elevation") return ""
					if (elements["heartrate"].hide  && seriesName === "heartrate") return ""
					if (elements["speed"].hide  && seriesName === "speed") return ""
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

		grid: ["speed", "elevation", "heartrate"].map((key) => ({
			left: window.innerWidth < 600 ? '36px' : '50px', 
			right: window.innerWidth < 600 ? '24px' : '40px', 
			top: ((elements as any)[key].top ?? 101) + '%', width: 'auto', 
			height: ((elements as any)[key].height ?? 0) + '%' 
		})),

		xAxis: [
			{
				show: !elements["speed"].hide,
				type: 'value',
				axisLabel: {
					formatter: formatDurationAxis,
				},
				max: data.Points[data.Points.length - 1].Duration / 60.0,
				gridIndex: 0,
			},
			{
				show: !elements["elevation"].hide,
				type: 'value',
				axisLabel: {
					formatter: formatDurationAxis,
				},
				max: data.Points[data.Points.length - 1].Duration / 60.0,
				gridIndex: 1,
			},
			{
				show: !elements["heartrate"].hide,
				type: 'value',
				axisLabel: {
					formatter: formatDurationAxis,
				},
				max: data.Points[data.Points.length - 1].Duration / 60.0,
				gridIndex: 2,
			},
		],

		yAxis: [
			{
				show: !elements["speed"].hide,
				type: 'value',
				scale: true,
				gridIndex: 0,
			},
			{
				show: !elements["elevation"].hide,
				type: 'value',
				scale: true,
				gridIndex: 1,
			},
			{
				show: !elements["heartrate"].hide,
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
				markLine: elements["speed"].hide ? undefined : {
					// Bug: cannot set symbol in data => disable also for average
					symbol: segmentMarker.length === 0 ? "arrow" : "none",
					data: [
						{ 
							type: 'average', 
							name: 'Avg', 
							label: {
								formatter: (val: { value: number }) => { 
									return val.value.toFixed(0) 
								},
							}
						},
						...segmentMarker
					],
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
					// Bug: cannot set symbol in data => disable also for average
					symbol: segmentMarker.length === 0 ? "arrow" : "none",
					data: [
						{ 
							type: 'average', 
							name: 'Avg', 
							label: { 
								formatter: (val: { value: number }) => { 
									return val.value.toFixed(0) 
								},
							}
						},
						...segmentMarker
					]
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
					// Bug: cannot set symbol in data => disable also for average
					symbol: segmentMarker.length === 0 ? "arrow" : "none",
					data: [
						{ 
							type: 'average', 
							name: 'Avg', 
							label: { 
								formatter: (val: { value: number }) => { 
									return val.value.toFixed(0) 
								},
							}
						},
						...segmentMarker
					]
				},
			},
		] as echarts.SeriesOption[]).concat(getHeartrateLines(data) as any) as any,

		// Support zooming without displaying toolbox: https://github.com/apache/echarts/issues/13397
		toolbox : {
			// Overlay any other label (like track segment)
			backgroundColor: darkTheme ? '#100c2a' : "#fff",
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
		setTimeout(() => chart.resize(), 400)
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