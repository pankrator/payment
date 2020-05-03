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
