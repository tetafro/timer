document.addEventListener('DOMContentLoaded', function () {
    if (window.location.pathname == "/") {
        setCurrentTime();
    } else if (window.location.pathname.startsWith("/timer")) {
        setTimer();
        setInterval(setTimer, 1000);
    }
});

// setCurrentTime set current time and timezone inputs on the main page.
function setCurrentTime() {
    let now = new Date();
    now.setMinutes(now.getMinutes() - now.getTimezoneOffset());

    document.getElementById("time").value = now.toISOString().substring(0, 16);
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

// zeroPadding adds leading zeros to the input value.
function zeroPadding(n, len) {
    let s = n.toString();
    for (let i = 0; i < (len - s.length); i++) {
        s = "0" + s;
    }
    return s;
}
