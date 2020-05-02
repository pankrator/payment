function createTransaction(amount, type, merchantUUID, customerEmail, dependsOnUUID, callback) {
    $.ajax({
        url: "/payment",
        method: "POST",
        data: JSON.stringify({
            "amount": parseInt(amount),
            "type": type,
            "merchant_id" :merchantUUID,
            "customer_email": customerEmail,
            "depends_on_uuid": dependsOnUUID,
        }),
        headers: {
            "X-CSRF-Token": loginData.csrfToken,
            "Authorization": "Bearer " + loginData.token,
            "Content-Type": "application/json"
        },
        success: (data, status, xhr) => {
            callback(data);
        },
        error: (err) => {
            console.log(err.responseText);
        }
    });
}

function transactionsOnload() {
    document.body.addEventListener("click", function(e) {
        if (e.target && e.target.id == "create") {
            let amount = document.getElementById("amount").value;
            let type = document.getElementById("type").value;
            let customerEmail = document.getElementById("customer-email").value;
            let dependsOnUUID = document.getElementById("depends-on").value;
            let merchantUUID = document.getElementById("merchant").value;

            createTransaction(amount, type, merchantUUID, customerEmail, dependsOnUUID, (data) => {
                console.log(data);
            });
        }
    });
}
