// Page init and timer updates.
document.addEventListener('DOMContentLoaded', function () {
    if (window.location.pathname == "/") {
        // document.getElementById("timezone").value = getCurrentTimezone();

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

    if (diff <= 0) {
        document.getElementById("days").innerText = "00";
        document.getElementById("hours").innerText = "00";
        document.getElementById("minutes").innerText = "00";
        document.getElementById("seconds").innerText = "00";
        document.getElementById("label-seconds").classList.add("alert");
        document.getElementById("seconds").classList.add("alert");
        document.getElementById("timeout").hidden = false;
        return;
    }

    const minute = 60;
    const hour = 60 * minute;
    const day = 24 * hour;

    let days = Math.trunc(diff / day);
    document.getElementById("days").innerText = zeroPadding(days, 2);
    diff -= days * day;

    let hours = Math.trunc(diff / hour);
    document.getElementById("hours").innerText = zeroPadding(hours, 2);
    diff -= hours * hour;

    let minutes = Math.trunc(diff / minute);
    document.getElementById("minutes").innerText = zeroPadding(minutes, 2);
    diff -= minutes * minute;

    document.getElementById("seconds").innerText = zeroPadding(diff, 2);
}

// zeroPadding adds leading zeros to the input value.
function zeroPadding(n, len) {
    let s = n.toString();
    for (let i = 0; i < (len - s.length); i++) {
        s = "0" + s;
    }
    return s;
}
