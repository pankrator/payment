let loginData = {};
let body;

window.onload = function() {
    body = this.document.getElementById("body");
    refreshToken(function() {
        if (!loginData.token) {
            loadView("/login", body, loginOnload)
        } else {
            loadView("/transactions", body, transactionsOnload)
        }
    })
}

function refreshToken(callback) {
    $.ajax({
        url: "/refresh",
        method: "GET",
        success: (data, status, xhr) => {
            let csrfToken = xhr.getResponseHeader("X-CSRF-Token");
            loginData.csrfToken = csrfToken;
            if (data) {
                loginData.token = data;
                callback();
            }
        },
        error: (err) => {
            callback();
        }
    })
}

function loadView(path, inElement, callback) {
    $.ajax({
        url: path,
        method: "GET",
        headers: {
            "Authorization": "Bearer " + loginData.token
        },
        success: (data, status, xhr) => {
            inElement.innerHTML = data;
            let csrfToken = xhr.getResponseHeader("X-CSRF-Token");
            loginData.csrfToken = csrfToken;
            callback();
        },
    });
}
