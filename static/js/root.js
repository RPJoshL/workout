document.addEventListener("DOMContentLoaded", () => {

})

document.addEventListener('htmx:afterRequest', function(evt) {

	// Check if we defined the notification container as a target
	const attr = evt.target.attributes
	const targetError = attr.getNamedItem("d-notification")
	if (targetError === null || targetError.value !== "error") {
		return
	}

	// We only show the notification if the request failed
	if (evt.detail.xhr.status < 400 && evt.detail.xhr.status != 0 ) {
		return
	}

	// Check whether we should use dark / light mode
	// const isDark = document.getElementById("dark") !== null
	const isDark = document.getElementById("content").classList.contains("theme-cust-dark")

	// Get the message
	let message = evt.detail.xhr.response
	if (message == "") {
		message = "Request failed with unknown reason"
	}

	// If we faced an internal error, render the full page because we don't
	// have any context anymore
	if (evt.detail.xhr.status === 500 && message.length > 1000) {
		document.write(message)
		return
	}

	const toastContent = document.createElement("span")
	toastContent.className = "content"

	// Icon component
	const icon = document.createElement("span")
	icon.className = "icon"
	icon.innerHTML = '<svg viewBox="0 0 24 24" width="100%" height="100%" fill="#e74c3c"><path d="M11.983 0a12.206 12.206 0 00-8.51 3.653A11.8 11.8 0 000 12.207 11.779 11.779 0 0011.8 24h.214A12.111 12.111 0 0024 11.791 11.766 11.766 0 0011.983 0zM10.5 16.542a1.476 1.476 0 011.449-1.53h.027a1.527 1.527 0 011.523 1.47 1.475 1.475 0 01-1.449 1.53h-.027a1.529 1.529 0 01-1.523-1.47zM11 12.5v-6a1 1 0 012 0v6a1 1 0 11-2 0z"></path></svg>'

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
		className: isDark ? "notification-error-dark" : "notification-error-light"
	}).showToast();
});

document.addEventListener('htmx:beforeRequest', function(evt) {
	evt.detail.xhr.setRequestHeader("Time-Zone", Intl.DateTimeFormat().resolvedOptions().timeZone)
})