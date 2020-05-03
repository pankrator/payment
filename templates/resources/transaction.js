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
            callback(null, data);
        },
        error: (err) => {
            callback(err);
        }
    });
}