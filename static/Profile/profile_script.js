document.addEventListener('DOMContentLoaded', function () {
    const login = document.getElementById('login');

    if (login) {
       login.addEventListener('click', function () {
            window.location.href = '/auth/battlenet';
        });
    }
});


function getCookie(name) {
    const cookies = document.cookie.split("; ");
    for (let cookie of cookies) {
        let [key, value] = cookie.split("=");
        if (key === name) return decodeURIComponent(value);
    }
    return null;
}

// Get userID from cookie
const userID = getCookie("userID");

if (userID) {
    console.log("User ID:", userID);

    // Update URL if userID is not already present
    const currentURL = new URL(window.location.href);
    if (!currentURL.pathname.includes(userID)) {
        window.history.replaceState(null, "", `/profile/${userID}`);
    }
} else {
    console.log("User not logged in");
    // Optionally redirect to login page
    window.location.href = "/";
}