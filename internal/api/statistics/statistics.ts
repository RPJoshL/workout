import * as echarts from "echarts";

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

export function AddListener() {
	// @ts-expect-error Declared globally
	OnElementReady("#workout-statistics", () => {
		const root = document.getElementById("workout-statistics")
		if (root === null) return

		const units = root.querySelectorAll(".unit")
		units.forEach((el) => {
			el.addEventListener("click", () => toggleUnit(el.getAttribute("data-id") ?? "", units))
		})

		const methods = root.querySelectorAll(".avg-sum")
		methods.forEach((el, i) => {
			el.addEventListener("click", () => toggleMethod(i, methods))
		})

		const workouts = root.querySelectorAll(".workout")
		workouts.forEach((el, idx) => {
			el.addEventListener("click", () => toggleWorkout(idx, workouts))
		})
	})
}


function toggleUnit(id: string, units: NodeListOf<Element>) {
	let oldId = ""
	units.forEach((unit) => {
		if (unit.getAttribute("data-selected") === "true") {
			oldId = unit.getAttribute("data-id") ?? ""
		}

		unit.setAttribute("data-selected", unit.getAttribute("data-id") == id ? "true" : "false")
	})

	if (oldId !== id) {
		// Query API
		console.log("Need to query API")
	}
}

function toggleMethod(idx: number, methods: NodeListOf<Element>) {
	const wasSelected = methods[idx].getAttribute("data-selected") === "true"

	methods.forEach((method, i) => {
		method.setAttribute("data-selected", idx === i ? "true" : "false")
	})

	if (!wasSelected) {
		console.log("Need to query API")
	}
}

function toggleWorkout(idx: number, workouts: NodeListOf<Element>) {
	let isOneWorkoutSelected = false

	let newSummaryState = workouts[0].getAttribute("data-selected") === "true"
	if (idx === 0) newSummaryState = !newSummaryState
	else if (newSummaryState) newSummaryState = false

	workouts.forEach((workout, i) => {
		let isSelected = workout.getAttribute("data-selected") === "true"

		if (i === idx || newSummaryState) {
			isSelected = !isSelected && !newSummaryState
			workout.setAttribute("data-selected", isSelected ? "true" : "false")
		}

		if (isSelected && i !== 0) {
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
		case (width > 2000): return 30;
		case (width > 1500): return 20;
		case (width > 1000): return 16;
		case (width > 600): return 12;
		default: return 10;
	}
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

	// No date selected already
	if (center == null || center == "") {
		const centerDate = new Date()
		centerDate.setDate(centerDate.getDate() - (count / 2))

		center = formatDate(centerDate)
		centerInput.setAttribute("data-date", center)
		centerInput.value = center
	}

	// Find the index of the center
	let centerIndex = -1
	const centerDate = parsePrettyDate(center)
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

	// Try to parse center date
	console.log("Counting with " + count + " with center index " + center + " found at " + centerIndex)

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

	return `${day}.${month}.${year}`
}

function parseDate(date: string): Date {
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
