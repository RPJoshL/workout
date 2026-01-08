import * as echarts from "echarts";
import { CallbackDataParams } from "echarts/types/dist/shared.js";

const SEARCH_FORM_ID = "workout-statistics-search"
const ALL_INDEX = 0
const CATEGORY_INDEX = 1

// The ID of "all" sampling unit
const UNIT_ALL_ID = "4"

export function renderChart(id: string, darkTheme: boolean, options: echarts.EChartsOption): echarts.ECharts {
	
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

export function AddSearchListener() {
	// @ts-expect-error Declared globally
	OnElementReady("#workout-statistics", () => {
		const root = document.getElementById("workout-statistics")
		if (root === null) return

		const units = root.querySelectorAll(".unit")
		units.forEach((el) => {
			el.addEventListener("click", () => toggleUnit(
				el.getAttribute("data-id") ?? "", 
				el.getAttribute("data-name") ?? "",
				units
			))
		})

		const methods = root.querySelectorAll(".avg-sum")
		methods.forEach((el, i) => {
			el.addEventListener("click", () => toggleMethod(i, methods))
		})

		const workouts = root.querySelectorAll(".workout")
		workouts.forEach((el, idx) => {
			el.addEventListener("click", () => toggleWorkout(idx, workouts))
			el.addEventListener("contextmenu", (ev) => {
				ev.preventDefault()
				toggleWorkout(idx, workouts, true)
			})
		})

		const form = document.getElementById("workout-statistics-search")
		form?.addEventListener("htmx:configRequest", (ev: any) => {
			// Add correct counts to request
			const count = getCountForUnit(getInputValue("samplingUnit"))
			ev.detail.parameters.displayCount = count
			ev.detail.parameters.count = count * 3

			// Add center date the user selected
			ev.detail.parameters.centerTime = (root.querySelector(".date-selection input") as HTMLInputElement | null)?.value
		})
	})
}

export function AddDateListener() {
	// @ts-expect-error Declared globally
	OnElementReady("#workout-statistics .date-selection", (root: HTMLDivElement) => {
		let lastTimeout = 0

		const setNewDate = (offsetCount: number) => {
			let centerDate = parsePrettyDate((root.querySelector("input"))?.value ?? "")	
			const cnt = getCountForResoulution()

			switch (getInputValue("samplingUnit").toUpperCase()) {
				case "":
				case "DAY": {
					centerDate.setDate(centerDate.getDate() + (cnt * offsetCount))
					break
				}
				case "WEEK": {
					centerDate.setDate(centerDate.getDate() + (cnt * offsetCount * 7))
					break
				}
				case "MON":
				case "MONTH": {
					centerDate.setMonth(centerDate.getMonth() + (cnt * offsetCount))
					break
				}
				case "YEAR": {
					centerDate.setFullYear(centerDate.getFullYear() + (cnt * offsetCount))
					break
				}
				case "ALL": {
					// Not supported for total
					return
				}
				default: {
					console.warn("Unknown sampling unit: " + getInputValue("samplingUnit"))
				}
			}

			// Don't allow to set values in the future. We don't have any data for this
			if (centerDate > (new Date())) {
				centerDate = new Date()
			}

			(root.querySelector("input") as any).value = formatDatePretty(centerDate)

			// Use a timeout to prevent multiple quick requests when navigating through the dates
			if (lastTimeout !== 0)  clearTimeout(lastTimeout)
			lastTimeout = setTimeout(() => {
				// @ts-expect-error // Globally declared
				htmx.trigger("#workout-statistics-search", "submit")
			}, 250)
		}

		root.querySelector(".left")?.addEventListener("click", () => {
			setNewDate(-1)
		})
		root.querySelector(".right")?.addEventListener("click", () => {
			setNewDate(+1)
		})

		// Refetch data when center date was updated
		const centerInput = document.querySelector("#workout-statistics .date-selection input.date") as HTMLInputElement | null
		centerInput?.addEventListener("keypress", (eve) => {
			if (eve.key !== "Enter" || centerInput.getAttribute("origin-value") === centerInput.value) {
				return
			}

			// @ts-expect-error // Globally declared
			htmx.trigger("#workout-statistics-search", "submit")
		})
		centerInput?.addEventListener("focusout", () => {
			if (centerInput.getAttribute("origin-value") === centerInput.value) {
				return
			}

			// @ts-expect-error // Globally declared
			htmx.trigger("#workout-statistics-search", "submit")
		})
	})
}

function toggleUnit(id: string, name: string, units: NodeListOf<Element>) {
	let oldId = ""
	units.forEach((unit) => {
		if (unit.getAttribute("data-selected") === "true") {
			oldId = unit.getAttribute("data-id") ?? ""
		}

		unit.setAttribute("data-selected", unit.getAttribute("data-id") == id ? "true" : "false")
	})

	if (oldId !== id) {
		setInputValue("samplingUnit", name)
		adjustDatePickerForUnit(oldId, id)

		// @ts-expect-error // Globally declared
		htmx.trigger("#workout-statistics-search", "submit")
	}
}

/** Adjusts the date picker to / from a range selection if the unit was changed to / from "all" */
function adjustDatePickerForUnit(old: string, neww: string) {
	if (old != UNIT_ALL_ID && neww != UNIT_ALL_ID) {
		return
	}

	const picker = document.getElementById(SEARCH_FORM_ID + "-date-picker")
	if (picker === null) return

	// @ts-expect-error flatpickr is set by library
	picker._flatpickr.set("mode", neww == UNIT_ALL_ID ? "range" : "single")
	
	// Reset any input when a range was selected for total
	if (old == UNIT_ALL_ID) {
		// @ts-expect-error flatpickr is set by library
		picker._flatpickr.clear()
	}
}

function setInputValue(name: string, value: string) {
	const search = document.getElementById(SEARCH_FORM_ID)
	if (search === null) return

	const input = search.querySelector('input[name="' + name + '"]')
	if (input === null) {
		console.warn("Didn't found an input with name: " + name)
		return
	}
	(input as any).value = value
}

function getInputValue(name: string): string {
	const search = document.getElementById(SEARCH_FORM_ID)
	if (search === null) return ""

	const input = search.querySelector('input[name="' + name + '"]')
	if (input === null) {
		console.warn("Didn't found an input with name: " + name)
		return ""
	}
	return (input as any).value
}

function toggleMethod(idx: number, methods: NodeListOf<Element>) {
	const wasSelected = methods[idx].getAttribute("data-selected") === "true"

	methods.forEach((method, i) => {
		method.setAttribute("data-selected", idx === i ? "true" : "false")
	})

	if (!wasSelected) {
		setInputValue("aggregation", methods[idx].getAttribute("data-name") ?? "")

		// @ts-expect-error // Globally declared
		htmx.trigger("#workout-statistics-search", "submit")
	}
}

function toggleWorkout(idx: number, workouts: NodeListOf<Element>, replaceSelection: boolean = false) {
	let isOneWorkoutSelected = false

	let newSummaryState = workouts[ALL_INDEX].getAttribute("data-selected") === "true"
	if (idx === 0) newSummaryState = !newSummaryState
	else if (newSummaryState) newSummaryState = false

	workouts.forEach((workout, i) => {
		let isSelected = workout.getAttribute("data-selected") === "true"

		if (i === idx || newSummaryState && i !== CATEGORY_INDEX) {
			isSelected = (!isSelected && !newSummaryState) || (i === idx && replaceSelection)
			workout.setAttribute("data-selected", isSelected ? "true" : "false")
		} else if (replaceSelection && i !== CATEGORY_INDEX) {
			isSelected = false
			workout.setAttribute("data-selected", "false")
		}

		if (isSelected && i !== ALL_INDEX && i !== CATEGORY_INDEX) {
			isOneWorkoutSelected = true
		}
	})

	// Toggle summary
	workouts[0].setAttribute("data-selected", isOneWorkoutSelected ? "false" : "true");

	// Update graph
	(window as any).workoutTypeSelectionChanged()
}

export type StatisticData = {
	start: string
	end: string
	label: string
	labelTooltip?: string
	id: number
}

export function getCountForResoulution(): number {
	const width = Math.max(
		document.body.scrollWidth,
		document.documentElement.scrollWidth,
		document.body.offsetWidth,
		document.documentElement.offsetWidth,
		document.documentElement.clientWidth
	);

	switch(true) {
		case (width > 2000): return 26;
		case (width > 1500): return 20;
		case (width > 1000): return 16;
		case (width > 600): return 12;
		default: return 10;
	}
}

export function getCountForUnit(unit: string) {
	let max = 100

	switch(unit.toUpperCase()) {
		case "DAY": { max = 100; break; }
		case "WEEK": { max = 30; break; }
		case "MON":
		case "MONTH": { max = 24; break; }
		case "YEAR": { max = 12; break; }
	}

	const countRes = getCountForResoulution()
	if (countRes > max) return max
	else return countRes
}

/**
 * Filters the provided data based on the clients resolution
 * and the selected center index
 */
export function filterData<Type extends StatisticData>(data: Type[]): Type[] {
	const count = getCountForResoulution()
	const centerInput = document.querySelector("#workout-statistics .date-selection input.date") as HTMLInputElement | null;
	let center = centerInput?.getAttribute("data-date")

	if (centerInput === null) return [];

	// No date selected already (only works for day)
	if (center == null || center == "") {
		const centerDate = new Date()
		centerDate.setDate(centerDate.getDate() - (count / 2))

		center = formatDate(centerDate)
		centerInput.setAttribute("data-date", center)
		centerInput.value = formatDatePretty(centerDate)
	}

	// Find the index of the center
	let centerIndex = -1
	const centerDate = parseDate(center)

	for(let i = 0; i < data.length; i++) {
		const d = data[i]

		if (centerDate >= parseDate(d.start) && centerDate <= parseDate(d.end)) {
			centerIndex = i
			break
		}
	}

	if (centerIndex === -1) {
		console.warn("Couldn't find center index for date '" + center + "' inside data")
		return []
	}

	let from = centerIndex - (count / 2)
	if (from < 0) from = 0
	let to = centerIndex + (count / 2) + 1 // slice excludes last element
	if (to > data.length) to = data.length

	return data.slice(from, to)
}

function formatDate(date: Date): string {
	const day = String(date.getDate()).padStart(2, "0")
	const month = String(date.getMonth() + 1).padStart(2, "0")
	const year = date.getFullYear()

	return `${year}-${month}-${day}`
}

export function formatDatePretty(date: Date): string {
	const day = String(date.getDate()).padStart(2, "0")
	const month = String(date.getMonth() + 1).padStart(2, "0")
	const year = date.getFullYear()

	return `${day}.${month}.${year}`
}

export function parseDate(date: string): Date {
	// ISO Times can correctly parsed without loosing time zone
	if (date.indexOf("T") !== -1 && !isNaN(Date.parse(date))) {
		return new Date(date)
	}

	// eslint-disable-next-line
	let [year, month, day] = date.split("-")
	if (year === undefined || month === undefined || day === undefined) {
		console.log("Unsupported date format: " + date)
	}

	let hour = 17
	let minute = 0
	if (day.indexOf("T") !== -1) {
		const times = day.substring(day.indexOf("T") + 1).split(":")
		day = day.substring(0, day.indexOf("T"))

		if (times.length >= 2) {
			hour = Number(times[0])
			minute = Number(times[1])
		}
	}

	return new Date(Number(year), Number(month) - 1, Number(day), hour, minute, 0)
}

function parsePrettyDate(date: string): Date {
	// eslint-disable-next-line
	let [day, month, year] = date.split(".")
	if (year === undefined || month === undefined || day === undefined) {
		console.log("Unsupported date format: " + date)
	}
	
	// Set to 16:00 to get the time zone correctly 
	return new Date(Number(year), Number(month) - 1, Number(day), 16, 0, 0)
}

/** 
 * Builds a tooltip HTML that looks like the default echarts tooltip base on the provided data. It's used to allow use to toggle
 * the name based on the input data
 */
export function buildTooltip(
	params: Array<CallbackDataParams>, data: Array<StatisticData>, 
	showName: boolean = true, showEmptyRowes: boolean = true,
	formatter: ((axisIndex: number, val: string|number|null) => string) | null = null
): string {
	let rtc = (params.length > 0 ? (data[params[0].dataIndex].labelTooltip ?? data[params[0].dataIndex].label) : "??") + "<br/>"
	
	rtc += `<table style='border-collapse:collapse;'>`
	params.forEach(p => {
		if (showEmptyRowes || (p.value !== null && p.value !== "" && p.value !== 0)) {
			const value = formatter ? formatter(p.componentIndex, p.value as any) : (p.value === null ? "" : (p.value as number).toLocaleString())
			rtc += `
				<tr>
					<td style='padding:2px 6px 2px 0;'>
						<span style='display:inline-block;width:12px;height:12px;border-radius:50%;background-color:${p.color};margin-right:4px'></span>
						<span>${showName ? p.seriesName : ""}</span>
					</td>
					<td style='padding:2px 0 2px 6px;text-align:right;'>
						<b>${value}</b>
					</td>
				</tr>`;
		}
	})
	rtc += `</table>`

	return rtc;
}

export function stepToNextHigherSamplingUnit(): boolean {
	const root = document.getElementById("workout-statistics")
	if (root === null) return false

	const units = root.querySelectorAll(".unit")
	units.forEach((el) => {
		el.addEventListener("click", () => toggleUnit(
			el.getAttribute("data-id") ?? "", 
			el.getAttribute("data-name") ?? "",
			units
		))
	})

	let idx = 0
	switch (getInputValue("samplingUnit").toUpperCase()) {
		case "":
		case "DAY": {
			idx = 0
			break
		}
		case "WEEK": {
			idx = 1
			break
		}
		case "MON":
		case "MONTH": {
			idx = 2
			break
		}
		case "YEAR": {
			idx = 3
			break
		}
		default: {
			console.warn("Unknown sampling unit: " + getInputValue("samplingUnit"))
			return false
		}
	}

	// Next lower unit
	idx = idx - 1
	if (idx < 0) return false

	toggleUnit(
		units[idx].getAttribute("data-id") ?? "", 
		units[idx].getAttribute("data-name") ?? "",
		units
	)

	return true
}

export function updateCenterDate(date: Date) {
	const input = document.querySelector("#workout-statistics .date-selection input")
	if (input === null) return;

	(input as HTMLInputElement).value = formatDatePretty(date)
}