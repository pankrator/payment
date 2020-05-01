function login(username, password, callback) {
    let request = $.ajax({
        url: "/login",
        method: "POST",
        headers: {
            "X-CSRF-Token": loginData.csrfToken,
            "Content-Type": "application/x-www-form-urlencoded"
        },
        data: "username="+username+"&password="+password,
        success: (token) => {
            callback(null, token);
        },
        error: (err) => {
            callback(err.responseText);
        }
    });
}

function loginOnload() {
    let loginButton = document.getElementById("login-button");
    loginButton.addEventListener("click", function() {
        let username = document.getElementById("username").value;
        let password = document.getElementById("password").value;
        login(username, password, function(err, token) {
            if (err) {
                let errBox = document.getElementById("login_err");
                errBox.innerHTML = "could not login " + err;
                return;
            }
            loginData.token = token;
            loadView("/transactions", body, function() {});
        });
    }, false);
}
