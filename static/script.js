// Page init and timer updates.
document.addEventListener('DOMContentLoaded', function () {
    if (window.location.pathname == "/") {
        setCurrentTime();

        // Timer buttons.
        document.getElementById("plus10m").addEventListener("click", () => {
            addTime(10, "minute");
        });
        document.getElementById("plus30m").addEventListener("click", () => {
            addTime(30, "minute");
        });
        document.getElementById("plus1h").addEventListener("click", () => {
            addTime(1, "hour");
        });
        document.getElementById("plus1d").addEventListener("click", () => {
            addTime(1, "day");
        });
    } else if (window.location.pathname.startsWith("/timer")) {
        setTimer();
        setInterval(setTimer, 1000);
    }
});

// setCurrentTime set current time and timezone inputs on the main page.
function setCurrentTime() {
    document.getElementById("time").value = timeToString(new Date());
    document.getElementById("timezone").value = getCurrentTimezone();
}

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

function addTime(duration, unit) {
    let timeInput = document.getElementById("time");
    let timezone = document.getElementById("timezone").value;
    let ts = Date.parse(timeInput.value+timezone);

    var time;
    switch (unit) {
        case "minute":
            time = new Date(ts + duration * 60 * 1000);
            break;
        case "hour":
            time = new Date(ts + duration * 60 * 60 * 1000);
            break;
        case "day":
            time = new Date(ts + duration * 24 * 60 * 60 * 1000);
            break;
        default:
            return;
    }

    timeInput.value = timeToString(time);
}

// getCurrentTimezone returns current timezone in "+00:00" format.
function getCurrentTimezone() {
    // Get offset in minutes
    let offset = -1 * (new Date()).getTimezoneOffset();

    let sign = "+";
    if (offset < 0) {
        sign = "-";
        offset *= -1;
    }

    // Offset in hours
    if (offset % 60 == 0) {
        offset /= 60;
        if (offset < 10) {
            return sign + "0" + offset.toString() + ":00";
        }
        return sign + offset.toString() + ":00";
    }

    // Offset in hours and minutes
    let minutes = offset % 60;
    let hr = zeroPadding((offset - minutes) / 60, 2);
    let min = zeroPadding(minutes, 2);

    return sign + hr + ":" + min;
}

// timeToString converts input time to a string keeping it in the same
// timezone, but not including the timezone itself.
function timeToString(t) {
    t.setMinutes(t.getMinutes() - t.getTimezoneOffset());
    return t.toISOString().substring(0, 16);
}

// zeroPadding adds leading zeros to the input value.
function zeroPadding(n, len) {
    let s = n.toString();
    for (let i = 0; i < (len - s.length); i++) {
        s = "0" + s;
    }
    return s;
}
