export function SetupCheckboxSelection(tableId: string) {
	// @ts-expect-error Declared globally
	OnElementReady("#" + tableId + " table", (table: HTMLTableElement) => {
		const allRows = table.querySelectorAll("tr")

		const headerCheckbox = table.querySelector("thead .select-checkbox input") as HTMLInputElement
		if (headerCheckbox) {
			headerCheckbox.addEventListener("change", () => {
				const isChecked = headerCheckbox.checked

				allRows.forEach((row) => {
					const input = row.querySelector(".select-checkbox input") as HTMLInputElement
					if (!input) return

					input.checked = isChecked
					if (isChecked) {
						row.classList.add("checkbox-selected")
					} else {
						row.classList.remove("checkbox-selected")
					}
				})
			})
		}

		allRows.forEach((row) => {
			const input = row.querySelector(".select-checkbox input") as HTMLInputElement
			if (!input) return

			// Remove selection when clicking on a selected row
			row.addEventListener("click", (e) => {
				// Don't do action if the checkbox itself was clicked
				if (e.target === input) return;

				if (input.checked) {
					input.checked = false
					row.classList.remove("checkbox-selected")

					e.preventDefault()
					e.stopPropagation()
				}
			})

			input.addEventListener("change", () => {
				if (input.checked) {
					row.classList.add("checkbox-selected")
				} else {
					row.classList.remove("checkbox-selected")

					// Make sure header checkbox is unchecked when a row is unchecked
					if (headerCheckbox) {
						headerCheckbox.checked = false
					}
				}
			})
		})
	})
}