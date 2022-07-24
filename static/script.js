document.addEventListener('DOMContentLoaded', function () {
    if (window.location.pathname == "/") {
        setCurrentTime();
    } else if (window.location.pathname.startsWith("/timer")) {
        formatTimer();
        setInterval(formatTimer, 1000);
    }
});

function setCurrentTime() {
    let now = new Date();
    now.setMinutes(now.getMinutes() - now.getTimezoneOffset());

    document.getElementById("time").value = now.toISOString().substring(0, 16);
    document.getElementById("timezone").value = getCurrentTimezone();
}

function formatTimer() {
    let deadline = parseInt(document.getElementById("deadline").innerText);
    let now = Math.trunc((new Date()) / 1000);
    let diff = deadline - now;

    if (diff <= 0) {
        document.getElementById("days").innerText = "00";
        document.getElementById("hours").innerText = "00";
        document.getElementById("minutes").innerText = "00";
        document.getElementById("seconds").innerText = "00";
        return;
    }

    const minute = 60;
    const hour = 60 * minute;
    const day = 24 * hour;

    if (diff >= day) {
        let days = Math.trunc(diff / day);
        document.getElementById("days").innerText = zeroPadding(days, 2);
        diff -= days * day;
    }
    if (diff >= hour) {
        let hours = Math.trunc(diff / hour);
        document.getElementById("hours").innerText = zeroPadding(hours, 2);
        diff -= hours * hour;
    }
    if (diff >= minute) {
        let minutes = Math.trunc(diff / minute);
        document.getElementById("minutes").innerText = zeroPadding(minutes, 2);
        diff -= minutes * minute;
    }
    if (diff >= 0) {
        document.getElementById("seconds").innerText = zeroPadding(diff, 2);
    }
}

function getCurrentTimezone() {
    let offset = -1 * (new Date()).getTimezoneOffset();

    let sign = "+";
    if (offset < 0) {
        sign = "-";
        offset *= -1;
    }

    if (offset % 60 == 0) {
        offset /= 60;
        if (offset < 10) {
            return sign + "0" + offset.toString() + ":00";
        }
        return sign + offset.toString() + ":00";
    }

    let minutes = offset % 60;
    let hours = (offset - minutes) / 60;
    if (hours < 10) {
        hours = "0" + hours.toString();
    } else {
        hours = hours.toString();
    }
    if (minutes < 10) {
        minutes = "0" + minutes.toString();
    } else {
        minutes = minutes.toString();
    }
    return sign + hours + ":" + minutes;
}

function zeroPadding(n, len) {
    let s = n.toString();
    for (let i = 0; i < (len - s.length); i++) {
        s = "0" + s;
    }
    return s;
}
