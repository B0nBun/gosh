const input = document.getElementById("url-input")
const form = document.getElementById("form")

function isURLValid(url) {
    try {
        new URL(url)
        return true
    } catch {
        return false
    }
}

form.addEventListener("submit", (e) => {
    const raw = input.value
    if (isURLValid(raw)) {
        form.dataset.label = ""
    } else {
        e.preventDefault()
        form.dataset.label = "Invalid URL"
    }
})

input.addEventListener("input", (e) => {
    const raw = e.currentTarget.value
    if (isURLValid(raw)) {
        form.dataset.label = ""
    }
})