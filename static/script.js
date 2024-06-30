// Page init and timer updates.
document.addEventListener('DOMContentLoaded', function () {
    if (window.location.pathname == "/") {
        const formDuration = document.getElementById("form-duration");
        formDuration.addEventListener("submit", function () {
            const inputDuration = formDuration.elements["duration"];
            const inputDeadline = formDuration.elements["deadline"];
            const current = new Date();
            const seconds = inputDuration.value * 60;
            const deadline = new Date(current.getTime() + seconds * 1000);
            inputDeadline.value = deadline.toISOString();
        });
        const formDeadline = document.getElementById("form-deadline");
        formDeadline.addEventListener("submit", function () {
            const inputCalendar = formDeadline.elements["calendar"];
            const inputDeadline = formDeadline.elements["deadline"];
            const deadline = new Date(inputCalendar.value);
            inputDeadline.value = deadline.toISOString();
        });
    } else if (window.location.pathname.startsWith("/timer")) {
        setTimer();
        setInterval(setTimer, 1000);
    }
});

// setTimer sets time to deadline on the timer page.
function setTimer() {
    let deadline = parseInt(document.getElementById("deadline").innerText);
    let now = Math.trunc((new Date()) / 1000);
    let diff = deadline - now;

    const timerParts = [
        {element: document.getElementById("days"), size: 60*60*24},
        {element: document.getElementById("hours"), size: 60*60},
        {element: document.getElementById("minutes"), size: 60},
        {element: document.getElementById("seconds"), size: 1},
    ];

    if (diff <= 0) {
        timerParts.forEach((part) => {
            setText(part.element, "00");
        });
        document.getElementById("label-seconds").classList.add("alert");
        document.getElementById("seconds").classList.add("alert");
        document.getElementById("timeout").hidden = false;
        return;
    }

    timerParts.forEach((part) => {
        const count = Math.trunc(diff / part.size);
        setText(part.element, zeroPadding(count, 2));
        diff -= count * part.size;
    });
}

function setText(element, text) {
    if (element != null) {
        element.innerText = text;
    }
}

// zeroPadding adds leading zeros to the input value.
function zeroPadding(n, len) {
    let s = n.toString();
    for (let i = 0; i < (len - s.length); i++) {
        s = "0" + s;
    }
    return s;
}
