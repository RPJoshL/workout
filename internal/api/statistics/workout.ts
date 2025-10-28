import { EChartsOption } from "echarts"
import * as echarts from "echarts";
import { buildTooltip, filterData, formatDatePretty, parseDate, renderChart, StatisticData, stepToNextHigherSamplingUnit, updateCenterDate } from "./statistics";
import { Dictionary } from "echarts/types/src/util/types.js";

const ID_ALL = -1
const ID_CATEGORIES = -2

export function InitWorkoutGraph(id: string, lang: "de" | "en", data: WorkoutData[], types: WorkoutType[], changeTabScriptName: string) {
	// @ts-expect-error Declared globally
	OnElementReady("#" + id, () => {
		createChart(id, lang, filterData(data), types, changeTabScriptName)
	})
}

interface WorkoutType {
	id: number
	nameDe: string
	nameEn: string
	tagDark: string
	tagWhite: string
}

interface WorkoutData extends StatisticData {
	distance: Record<number, number>
	calories: Record<number, number>
	duration: Record<number, number>
	pai: Record<number, number>
	count: Record<number, number>
	speed: Record<number, number>
	heartrate: Record<number, number>
}

type typeOptions = {
	color: string
	translation: {
		de: string
		en: string
	}
	axisFormatter?: string | ((max: number) => ((val: any) => string))
	tooltipFormatter?: (val: number) => string
	dataKey: string
}

type typeKeys = "distance" | "calories" | "duration" | "speed" | "pai" | "heartRate"
const types: Record<typeKeys, typeOptions> = {
	"distance": {
		color: "#5070dd",
		translation: {
			de: "Distanz",
			en: "Distance"
		},
		axisFormatter: (max: number) => (dist: number) => {
			if (max > 10_000) return Math.round(dist / 1000).toLocaleString() + " km"
			else if (max > 1_000) return (Math.round(dist / 100) / 10).toLocaleString(undefined, { minimumFractionDigits: 1, maximumFractionDigits: 1 }) + " km"
			else return dist.toFixed(0) + " m"
		},
		tooltipFormatter(dist) {
			if (dist > 100_000) return Math.round(dist / 1000).toLocaleString() + " km"
			if (dist > 5_000) return (Math.round(dist / 100) / 10).toLocaleString(undefined, { minimumFractionDigits: 1, maximumFractionDigits: 1 }) + " km"
			else if (dist > 1_000) return (Math.round(dist / 10) / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) + " km"
			return dist.toFixed(0) + " m"
		},
		dataKey: "distance"
	},
	"calories": {
		color: "#ed8200",
		translation: {
			de: "Kalorien",
			en: "Calories"
		},
		axisFormatter: '{value} cal',
		dataKey: "calories",
		tooltipFormatter: (val) => val.toLocaleString() + " cal"
	},
	"duration": {
		color: "#b3a594",
		translation: {
			de: "Dauer",
			en: "Dauer"
		},
		axisFormatter: () => (sec: number) => {
			const hours = Math.floor(sec / 3600);
			const minutes = Math.floor((sec % 3600) / 60);
			return hours + ":" + (minutes < 10 ? "0" : "") + minutes
		},
		tooltipFormatter: (sec: number) => {
			const hours = Math.floor(sec / 3600);
			const minutes = Math.floor((sec % 3600) / 60);
			return hours + ":" + (minutes < 10 ? "0" : "") + minutes
		},
		dataKey: "duration"
	},
	"speed": {
		color: "#52eb00",
		translation: {
			de: "Geschwindigkeit",
			en: "Speed"
		},
		axisFormatter: (max) => (speed) => {
			// Format lower speeds as min/km
			if (max < 12) {
				speed = 3600 / speed
				const minutes = Math.floor(speed / 60);
				const seconds = Math.round(speed % 60);
				return `${minutes}:${seconds.toString().padStart(2, "0")} min/km`;
			} else {
				return speed.toLocaleString(undefined, { minimumFractionDigits: 1, maximumFractionDigits: 1 }) + " km/h"
			}
		},
		tooltipFormatter: (speed) => {
			if (speed <= 0) return "0"

			// Format lower speeds as min/km
			if (speed < 12) {
				speed = 3600 / speed
				const minutes = Math.floor(speed / 60);
				const seconds = Math.round(speed % 60);
				return `${minutes}:${seconds.toString().padStart(2, "0")}`;
			} else {
				return speed.toLocaleString(undefined, { minimumFractionDigits: 1, maximumFractionDigits: 1 }) + " km/h"
			}
		},
		dataKey: "speed"
	},
	"pai": {
		color: "#00eb89",
		translation: {
			de: "PAI",
			en: "PAI"
		},
		dataKey: "pai"
	},
	"heartRate": {
		color: "#9500c2",
		translation: {
			de: "Puls",
			en: "Heart rate"
		},
		axisFormatter: '{value} bpm',
		dataKey: "heartrate",
		tooltipFormatter: (val) => val.toFixed(0) + " bpm"
	},
}

function createChart(id: string, lang: "de" | "en", data: WorkoutData[], workoutTypes: WorkoutType[], changeTabScriptName: string) {
	/** The amount of data series we do have (distance, calories, etc...) */
	const seriesTypesCnt = 6

	const yAxis = Object.values(types).map((opt, idx) => ({
		type: "value",
		name: opt.translation[lang],
		axisLabel: {
			formatter: typeof opt.axisFormatter === "function" ? opt.axisFormatter(getMax([-1], opt.dataKey, data as any)) : opt.axisFormatter
		},
		axisLine: {
			show: true,
			lineStyle: {
				color: opt.color
			},
		},
		show: idx === 0
	} as echarts.YAXisComponentOption ))
	yAxis.push(...workoutTypes.map((t) => ({
		type: "value",
		name: t.id.toString(),
		axisLabel: {
			show: true,
		},
		axisLine: {
			show: false,
			lineStyle: {
				color: t.tagDark
			}
		},
		show: false // We never want to show this axis. We only want the data
	} as echarts.YAXisComponentOption )))

	const series = Object.values(types).map((opt, idx) => {
		const seriesData: number[] = data.map(d => {
			const dd = ((d as any)[opt.dataKey] as any) as Record<number, number>
			if (dd === undefined) {
				console.warn("Found no data key for: " + opt.dataKey)
				return 0
			}

			return dd[-1]
		})

		return {
			name: opt.translation[lang],
			type: "bar",
			data: seriesData,
			itemStyle: {
				color: opt.color
			},
			yAxisIndex: idx,
		} as echarts.SeriesOption})
	series.push(...workoutTypes.map((t, idx) => ({
		name: lang === "de" ? t.nameDe : t.nameEn,
		type: "bar",
		data: [], // Filled dynamically, because we need to know which data (distance, calories, etc.) to use
		itemStyle: {
			color: t.tagDark
		},
		emphasis: {
			focus: 'series'
		},
		yAxisIndex: seriesTypesCnt + idx,
	} as echarts.SeriesOption)))

	let axisIndexMultipleTypes = 0
	const options: EChartsOption = {
		legend: {
			data: Object.values(types).map((opt,) => ({
				name: opt.translation[lang],
				itemStyle: { color: opt.color }
			})),
			selected: {
				...Object.fromEntries(
					Object.entries(types).map(([_, val], idx) => [val.translation[lang], idx === 0])
				),
				...Object.fromEntries(workoutTypes.map(t => [lang === "de" ? t.nameDe : t.nameEn, false]))
			},
		},
		tooltip: {
			trigger: 'axis',
			axisPointer: {
				type: 'cross',
				label: {
					formatter: (p) => {
						if (p.value === null) return ""
						if (typeof p.value !== "number") return p.value.toString()

						const opt = Object.values(types)[p.axisIndex >= seriesTypesCnt ? axisIndexMultipleTypes : p.axisIndex]
						if (opt.tooltipFormatter) {
							return opt.tooltipFormatter(p.value as number)
						}

						return p.value.toString()
					}
				},
			},
			formatter: (params) => buildTooltip(params as any, data, true, false, (axisIndex, val) => {
				const opt = Object.values(types)[axisIndex >= seriesTypesCnt ? axisIndexMultipleTypes : axisIndex]
				if (opt.tooltipFormatter) {
					return opt.tooltipFormatter(val as number)
				} else {
					return (val === null ? "" : (val as number).toLocaleString())
				}
			})
		},
		xAxis: [
			{
				type: "category",
				data: data.map((d) => d.label)
			}
		],
		yAxis: yAxis,
		series: series,
	}

	const chart = renderChart(id, true, options)

	let oldActiveAxes: number[] = [0]
	const adjustChart = (selected: Record<string, boolean>, activeAxes: number[] = []) => {
		// Get indexes of all active axis
		if (activeAxes.length === 0) {
			(options.series as echarts.SeriesOption[]).forEach((s, i) => {
				if (selected[s.name!]) {
					activeAxes.push(i)
				}
			})
		}

		// Check if categories should be shown instead of single bars / grouping them
		const selectedTypesRaw = getSelectedTypes()
		const selectedTypes = getSelectedTypesData(workoutTypes)
		const showCategories = selectedTypesRaw.find(i => i === ID_CATEGORIES) === ID_CATEGORIES
		const showAll = selectedTypesRaw.find(i => i === ID_ALL) === ID_ALL || selectedTypes.length === 0

		// Only allow a single selection when multiple workout types are selected
		// to not blow up the chart
		if ( (selectedTypes.length > 1 || showCategories)  && activeAxes.length > 1) {
			// Toggle to the new axis
			let newActiveIndex = 0

			// Find the new selected index and use it (and deselect the old one)
			const newSelection = activeAxes.filter((idx) => oldActiveAxes.indexOf(idx) === -1)
			if (newSelection.length > 0) newActiveIndex = newSelection[0]
			else newActiveIndex = 0

			activeAxes = [newActiveIndex]
		}
		// Deselect (of all types) is not possible when multiple types are selected
		if (selectedTypes.length > 1 && activeAxes.length === 0) {
			activeAxes = oldActiveAxes
		}

		// Store for refresh
		(window as any).lastWorkoutActiveAxes = activeAxes

		// Adjust offset and position
		const offset = 75
		let leftOffset = offset * -1
		let rightOffset = offset * -1
		const newYAxis = (options.yAxis as echarts.YAXisComponentOption[]).map((axis, i) => {
			if (activeAxes.includes(i)) {
				const idx = activeAxes.indexOf(i)

				// Adjust max values
				const opt = Object.values(types)[i]
				axis.axisLabel = {
					formatter: typeof opt.axisFormatter === "function" ? 
						opt.axisFormatter(getMax(
							showAll ? [-1] : selectedTypes.map(t => t.id), 
							opt.dataKey, data as any
						)) : 
						opt.axisFormatter
				}

				if (idx == 0 || idx < (activeAxes.length - 1) / 2) {
					leftOffset += offset
					return { ...axis, offset: leftOffset, position: 'left', show: true, max: undefined } as echarts.YAXisComponentOption
				} else {
					rightOffset += offset
					return { ...axis, offset: rightOffset, position: 'right', show: true, max: undefined }
				}
			} else {
				return { ...axis, offset: 0, show: false, max: undefined }
			}
		})

		const newLegendSelection = ((options.legend as echarts.LegendComponentOption).selected as Dictionary<boolean>)
		Object.keys(newLegendSelection).forEach((key, i) => {
			const yAxisIndex = (options.series as echarts.SeriesOption[]).findIndex(s => s.name === key)
			const workoutTypeIndex = selectedTypes.find(t => (lang === "de" ? t.nameDe : t.nameEn) === key)

			newLegendSelection[key] = activeAxes.indexOf(yAxisIndex) !== -1 || (showCategories && showAll && i >= seriesTypesCnt) || (workoutTypeIndex !== undefined && selectedTypes.length > 1)
		})

		let max = 0
		let applyMax = false
		const newSeries = (options.series as echarts.SeriesOption[]).map((s, i) => {
			// Default selection without any selected workout type
			if (!showCategories && showAll) {
				return s
			}

			// A single workout type is selected. We support showing all different series here.
			// => Modify series data for this specific type
			if (selectedTypes.length === 1) {
				// Not showing individual workout types
				if (i >= seriesTypesCnt) {
					return s
				}

				const dataKey = Object.values(types)[i].dataKey
				const workoutType = selectedTypes[0]

				return {
					...s,			
					stack: null,	
					data: data.map(d => {
						const dd = ((d as any)[dataKey] as any) as Record<number, number>
						if (dd === undefined) {
							console.warn("Found no data key for: " + dataKey)
							return 0
						}
			
						if(dd[workoutType.id] !== undefined) {
							return dd[workoutType.id]
						}

						return 0
					})
				}
			}

			// Multiple workout types are selected. We don't show the series
			if (i < seriesTypesCnt) {
				return {
					...s,
					stack: null,
					// To use this as a total value in popup
					type: activeAxes[0] === i  ? "custom" : "bar",
					itemStyle: {
						color: 'transparent'
					},
				}
			}

			const typ = Object.values(types)[activeAxes[0]]
			axisIndexMultipleTypes = activeAxes[0]
			let workoutType = selectedTypes.find(t => t.id === workoutTypes[i - seriesTypesCnt].id)

			// We have to use all workout categories for showing a stack
			if (workoutType === undefined && showCategories && showAll) {
				workoutType = workoutTypes[i - seriesTypesCnt]
			}
			if (workoutType === undefined) return s

			applyMax = true
			return {
				...s,
				stack: showCategories ? "a" : null,
				data: data.map(d => {
					const dd = ((d as any)[typ.dataKey] as any) as Record<number, number>
					if (dd === undefined) {
						console.warn("Found no data key for: " + typ.dataKey)
						return 0
					}
			
					if(dd[workoutType.id] !== undefined) {
						if (dd[workoutType.id] > max) max = dd[workoutType.id]

						return dd[workoutType.id]
					}

					return 0
				}),
			}
		})

		// Apply max value
		if (applyMax) {
			let newMax = Math.ceil(max * 1.05)
			if (newMax > 10000) newMax = Math.round(max / 1000) * 1000
			else if (newMax > 1000) newMax = Math.round(max / 100) * 100
			else if (newMax > 100) newMax = Math.round(max / 10) * 10

			for (let i = 0; i < newYAxis.length; i++) {
				if (i == activeAxes[0] || i >= seriesTypesCnt) {
					newYAxis[i] = {
						...(newYAxis[i] as echarts.YAXisComponentOption),
						max: newMax
					} as any
				}
			}
		}

		chart.setOption(
			{
				yAxis: newYAxis,
				legend: {
					...(options.legend as echarts.LegendComponentOption),
					selected: newLegendSelection
				},
				series: newSeries,
			},
			{
				replaceMerge: [ 'series' ]
			}
		)

		oldActiveAxes = activeAxes
	}

	chart.on('legendselectchanged', (p: any) => adjustChart(p.selected));
	
	// Add callback for workout type selection
	(window as any).workoutTypeSelectionChanged = () => adjustChart({}, oldActiveAxes)

	// Initial adjust
	adjustChart({}, (window as any).lastWorkoutActiveAxes ?? [0])

	// Add click listener for bar
	let lastTap = 0;
	chart.on('click', (ev) => {
		// No hoover support on mobile => user should use a double click on mobile
		if (navigator.maxTouchPoints === 0 ) {
			handleGraphZoom(parseDate(data[ev.dataIndex].start), parseDate(data[ev.dataIndex].end), changeTabScriptName)
		} else {
			const currentTime = new Date().getTime();
			const tapLength = currentTime - lastTap;

			// Double tap detected
			if (tapLength < 200 && tapLength > 0) {
				handleGraphZoom(parseDate(data[ev.dataIndex].start), parseDate(data[ev.dataIndex].end), changeTabScriptName)
			}

			lastTap = currentTime;
		}
	})
	// Double click event won't be fired with touch events. Searching for range is therefore not supported
	chart.on('dblclick', (ev) => {
		handleWorkoutSearch(parseDate(data[ev.dataIndex].start), parseDate(data[ev.dataIndex].end), changeTabScriptName)
	})
}

function handleGraphZoom(start: Date, end: Date, changeTabScriptName: string) {
	const middle = new Date((start.getTime() + end.getTime()) / 2)

	// No zoom in possible => open workout search overview
	const diffDays = (end.getTime() - middle.getTime()) / (1000.0 * 60 * 60 * 24)
	if (diffDays < 1) {
		handleWorkoutSearch(start, end, changeTabScriptName)
		return
	}

	// Adjust values
	updateCenterDate(middle)
	stepToNextHigherSamplingUnit()
}

function handleWorkoutSearch(start: Date, end: Date, changeTabScriptName: string) {
	const params = new URLSearchParams({
		dateRange: `${formatDatePretty(start)} to ${formatDatePretty(end)}`,
	})
	getSelectedTypes().filter(id => id !== ID_ALL && id !== ID_CATEGORIES).forEach(typ => params.append("types", typ.toString()))

	eval(`${changeTabScriptName}(3, false, false)`)

	// @ts-expect-error Declared globally
	htmx.ajax('GET', '/workout/?' + params.toString(), {
		target: '#content',
		swap: 'outerHTML transition:true'
	});
}

function getSelectedTypes(): number[] {
	const workoutTypes: number[] = []

	document.querySelectorAll("#workout-statistics .action-container .workout").forEach((el) => {
		if (el.getAttribute("data-selected") === "true") {
			workoutTypes.push(Number(el.getAttribute("data-id")))
		}
	})

	return workoutTypes
}

function getSelectedTypesData(data: WorkoutType[]): WorkoutType[] {
	const selectedTypes = getSelectedTypes()
	if (selectedTypes.length === 0) return []

	return selectedTypes
		.filter((t) => t !== ID_ALL && t !== ID_CATEGORIES)
		.map((t) => {
			// Not required anymore but for fallback
			if (t === ID_ALL|| t === ID_CATEGORIES) return {
				id: t,
				nameDe: "Gesamt",
				nameEn: "Total",
				tagDark: "#000000",
				tagWhite: "#ffffff"
			} as WorkoutType
		
			return data.find(dt => dt.id === t) as WorkoutType
		})
}

function getMax(workoutType: number[], key: string, data: Record<string, Record<string, number>>[]): number {
	let max = 0

	data.forEach((row) => {
		workoutType.forEach((typ) => {
			if ((row[key][typ] ?? 0) > max) {
				max = row[key][typ]
			}
		})
	})

	return max
}
