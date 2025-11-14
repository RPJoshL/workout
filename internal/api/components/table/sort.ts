export function SetupSortListener(id: string, defaultIdx: number) {
	// @ts-expect-error Declared globally
	OnElementReady("#" + id + " table", (table: HTMLTableElement) => {
		const allElements = table.querySelectorAll("th")
		allElements.forEach((el, idx: number) => {
			setSortListener(el, table, idx, defaultIdx, allElements as any)
		})
	})
}

function setSortListener(el: HTMLElement, root: HTMLTableElement, idx: number, defaultIdx: number, allElements: HTMLElement[]) {
	el.addEventListener("click", () => {
		// Deselect sorting mode of all other columns
		allElements.forEach(ae => {
			if (ae !== el && ae.getAttribute("data-sortDirection") !== "None") {
				ae.setAttribute("data-sortDirection", "None")
			}
		})

		const tbody = root.querySelector('tbody')
		if (!tbody) return

		const allRows = Array.from(tbody.querySelectorAll('tr'))
		if (allRows.length === 0) return

		// First row should not be sorted (sticky header for border)
		const firstRow = allRows[0]
		const rowsToSort = allRows.slice(1)

		let newDirection = "Asc"
		switch(el.getAttribute("data-sortDirection")) {
			case "Asc": {
				newDirection = "Desc"
				break
			}
			case "Desc": {
				newDirection = "None"
				break
			}
			case "None": {
				newDirection = "Asc"
				break
			}
		}

		// Reset index to date (default compare)
		let sortCoumnIdx = idx
		if (newDirection === "None") {
			sortCoumnIdx = defaultIdx
		}

		rowsToSort.sort((a, b) => {
			const valA = getCompareValue(a.children[sortCoumnIdx] as any, el)
			const valB = getCompareValue(b.children[sortCoumnIdx] as any, el)

			if (typeof valA === "number" && typeof valB === "number") {
				return newDirection === "Asc" ? valA - valB : valB - valA;
			}

			return newDirection === "Asc" ? valA.localeCompare(valB) : valB.localeCompare(valA)
		})

		// Reihenfolge im DOM aktualisieren
		tbody.innerHTML = ""
		tbody.appendChild(firstRow)
		rowsToSort.forEach(row => tbody.appendChild(row))
		el.setAttribute("data-sortDirection", newDirection)
	})
}

function getCompareValue(el: HTMLElement, header: HTMLElement): any {
	let val = el.innerText as any
	if (el.hasAttribute("data-compare-value")) {
		val = el.getAttribute("data-compare-value")
	}

	if (val === "") val = null
	else if (val === undefined) val = null

	// Parse as number
	if (header.getAttribute("data-type-number") === "true") {
		val = parseFloat(val)
		if (isNaN(val)) val = 0
	}

	return val
}