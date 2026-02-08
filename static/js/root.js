document.addEventListener("DOMContentLoaded", () => {
	let startY = 0;
	let isPulling = false;
	let pullD = 0
	const pullToRefresh = document.getElementById('pullToRefresh');
	const spinner = pullToRefresh.getElementsByClassName("spinner")[0]
	
	document.addEventListener('touchstart', (e) => {
		const elements = [
			document.querySelector("#workout-page-overview .workout-info-list"),
			document.querySelector("#workout-page-overview #workout-overview-search"),
			document.querySelector("#dashboard-page")
		]

		// Ignore scrolling from special containers
		const ignored = [ "leaflet-container", "custom-select" ]
		let isIgnored = false
		const walkTree = (element, index) => {
			ignored.forEach(ig => { if(element.classList.contains(ig)) isIgnored = true} )
			if (isIgnored || index > 10) return

			if (element.parentElement) {
				walkTree(element.parentElement, index++)
			}
		}
		walkTree(e.target)
		if (isIgnored) return

		// Check modal content only if it's visible
		const modalContent = document.getElementById("modal-content")
		if (modalContent.getAttribute("data-visible") === "true" && modalContent.scrollTop !== 0) return

		// We only want to scroll if scroll position of main container is zero
		const scrollAvailable = elements.filter((e) => e !== null && e !== undefined)
		// All has to be zero
		const scrollAvailableZero = scrollAvailable.filter(e => e.scrollTop === 0)

		if (scrollAvailable.length > 0 && scrollAvailable.length === scrollAvailableZero.length) {
			startY = e.touches[0].pageY;
			isPulling = true;
			pullD = 0
		}
	});
	
	document.addEventListener('touchmove', (e) => {
		if (isPulling) {
			e.preventDefault()
			const y = e.touches[0].pageY;
			if (y > startY) {
				const pullDistance = y - startY;
				let clampedDistance = Math.min(pullDistance, 100); // Limit the pull distance to 100px
				pullD = pullDistance

				if (clampedDistance >= 100) {
					clampedDistance = 100
					pullToRefresh.style.opacity = "1"
				} else {
					spinner.style.animation = "none"
					pullToRefresh.style.opacity = ".4"
				}

				pullToRefresh.style.setProperty('--pull-top', `${clampedDistance - 50}px`);
			}
		}
	});
	
	document.addEventListener('touchend', () => {
		if (isPulling) {
			if (pullD >= 100) {
				pullToRefresh.style.setProperty('--pull-top', `50px`);
				spinner.style.animation = "spin 1s linear infinite"
				console.log("Reloding page because of pulling down")
				
				if (window.navigation) window.navigation.reload()
				else window.location.reload()
			} else {
				pullToRefresh.style.setProperty('--pull-top', `-50px`);
				spinner.style.animation = "none"
			}
			isPulling = false;
		}
	});
}, { once: true })

document.addEventListener('htmx:afterRequest', function(evt) {

	// Get attributes of target element
	const attr = evt.target.attributes

	// Check if we defined the notification container as a target
	const targetError = attr.getNamedItem("d-notification")
	if (targetError === null) return
	const showError = targetError.value.includes("error")
	const showSuccess = targetError.value.includes("success")
	if (!showError && !showSuccess) return

	const xhr = evt.detail.xhr

	// We only show the notification if the request failed
	let isError = true
	if (xhr.status < 400 && evt.detail.xhr.status != 0 && !showSuccess ) {
		return
	} else if (xhr.status >= 200 && evt.detail.xhr.status < 300 && showSuccess) {
		isError = false
	}

	// Get the message
	let message = xhr.response
	if (message == "") {
		if (isError) message = "Request failed with unknown reason"
		else         message = "Success"
	}
	// When htmx:abort is used, the status is 0 and response is empty
	if (xhr.status === 0 && xhr.responseText === "") {
		console.info("Request was aborted, no notification will be shown")
		return
	}

	// If we faced an internal error, render the full page because we don't
	// have any context anymore
	if (xhr.status === 500 && message.length > 1000) {
		document.write(message)
		return
	}

	// Reauthentication popup should be visible
	if (xhr.status === 403 && !evt.detail.pathInfo.requestPath.includes("/login") ) {
		// Get action to perform
		const onReauthenticate = attr.getNamedItem("hx-on::reauthenticate")
		if (onReauthenticate) {
			console.log(onReauthenticate)
			eval(onReauthenticate.nodeValue)

			// Disable further processing of event
			evt.preventDefault()
			return
		}
	}

	// Show notification
	notify(message, isError)

	// Disable further processing of event
	evt.preventDefault()

	// AfterRequest scripts should still be called
	const afterRequest = attr.getNamedItem("hx-on::after-request")
	if (afterRequest && afterRequest.value !== "") eval(afterRequest.value)
});

document.addEventListener('htmx:beforeRequest', function(evt) {
	evt.detail.xhr.setRequestHeader("Time-Zone", Intl.DateTimeFormat().resolvedOptions().timeZone)
})

// eslint-disable-next-line @typescript-eslint/no-unused-vars
function ClosePopup() {
	const modal = document.getElementById("popup-root-wrapper")
	modal.setAttribute("data-visible", "false")
	setTimeout(() => {
		modal.setAttribute("data-visible-delayed", "false")
		modal.classList.remove()
	}, 450)
}

function AddTooltipListener() {
	const onTooltipClick = (element, e) => {
		e.stopPropagation()
		element.classList.add("hover")

		// Block any clicks
		const clickBlocker = document.getElementById("click-blocker")
		clickBlocker.setAttribute("data-visible", "true")
		
		// Remove the class when clicking anywhere other
		const otherClickListener = () => {
			clickBlocker.setAttribute("data-visible", "false")
			document.removeEventListener("click", otherClickListener)

			element.classList.remove("hover")
		}
		document.addEventListener("click", otherClickListener)
	}

	document.querySelectorAll("[data-tooltip]").forEach((el) => {
		// Already attached
		if (el.getAttribute("tooltip-listener") === "true") return
		
		// Set attribute for listener
		el.setAttribute("tooltip-listener", "true")

		el.addEventListener("click", (e) => onTooltipClick(el, e))
	})
}
document.addEventListener("DOMContentLoaded", () => AddTooltipListener());

// eslint-disable-next-line @typescript-eslint/no-unused-vars
function notify(message, isError) {
	// Check whether we should use dark / light mode
	// const isDark = document.getElementById("dark") !== null
	const content = document.getElementById("content")
	let isDark = content.classList.contains("theme-cust-dark")

	// If the content contains no theme specification, try to get it by a child div
	if (!isDark && !content.classList.contains("theme-cust-light")) {
		isDark = content.querySelectorAll(".theme-cust-dark").length > 0
	}

	const toastContent = document.createElement("span")
	toastContent.className = "content"

	// Icon component
	const icon = document.createElement("span")
	icon.className = "icon"
	icon.innerHTML = '<svg viewBox="0 0 24 24" width="100%" height="100%" fill="#e74c3c"><path d="M11.983 0a12.206 12.206 0 00-8.51 3.653A11.8 11.8 0 000 12.207 11.779 11.779 0 0011.8 24h.214A12.111 12.111 0 0024 11.791 11.766 11.766 0 0011.983 0zM10.5 16.542a1.476 1.476 0 011.449-1.53h.027a1.527 1.527 0 011.523 1.47 1.475 1.475 0 01-1.449 1.53h-.027a1.529 1.529 0 01-1.523-1.47zM11 12.5v-6a1 1 0 012 0v6a1 1 0 11-2 0z"></path></svg>'
	if (!isError) icon.innerHTML = '<svg viewBox="0 0 24 24" fill="#4aa850" xmlns="http://www.w3.org/2000/svg">    <path d="M0 0h24v24H0z" fill="none"/>    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/></svg>'

	// Text component
	const text = document.createElement("span")
	text.className = "text"
	text.appendChild(document.createTextNode(message))

	toastContent.appendChild(icon)
	toastContent.appendChild(text)

	// eslint-disable-next-line no-undef
	Toastify({
		node: toastContent,
		duration: 5000,
		newWindow: true,
		close: true,
		gravity: "top", // `top` or `bottom`
		position: "right", // `left`, `center` or `right`
		stopOnFocus: true, // Prevents dismissing of toast on hover
		onClick: function(){},
		progressBar: true,
		progressBarPosition: 'bottom',
		className: `notification-${isError ? "error" : "success"}-${isDark ? "dark" : "light"}`
	}).showToast();
}