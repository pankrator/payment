let loginData = {};
let body;

window.onload = function() {
    body = this.document.getElementById("body");
    init();
    refreshToken(function() {
        if (!loginData.token) {
            loadView("/login", body)
        } else {
            loadView("/transactions", body)
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
        },
    });
}


function init() {
    document.body.addEventListener("click", function(e) {
        if (e.target && e.target.id == "login-button") {
            let username = document.getElementById("username").value;
            let password = document.getElementById("password").value;
            login(username, password, function(err, token) {
                if (err) {
                    let errBox = document.getElementById("login_err");
                    errBox.innerHTML = "could not login " + err;
                    return;
                }
                loginData.token = token;
                loadView("/transactions", body);
            });
        }
    }, false);

    document.body.addEventListener("click", function(e) {
        if (e.target && e.target.id == "create") {
            let amount = document.getElementById("amount").value;
            let type = document.getElementById("type").value;
            let customerEmail = document.getElementById("customer-email").value;
            let dependsOnUUID = document.getElementById("depends-on").value;
            let merchantUUID = document.getElementById("merchant").value;

            createTransaction(amount, type, merchantUUID, customerEmail, dependsOnUUID, (err, data) => {
                if (err) {
                    let errBox = document.getElementById("transaction-error-box");
                    errBox.innerHTML = err.responseText;
                    return;
                }
                loadView("/transactions", body);
            });
        }
    });
}