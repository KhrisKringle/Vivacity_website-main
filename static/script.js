const get_user = ()=>{
    return fetch("/auth/user").then(res=>res.json())
}

document.addEventListener('DOMContentLoaded', function () {
    const kingpinsButton = document.getElementById('kingpinsButton');
    const vivacityButton = document.getElementById('vivacityButton');
    const vagrantsButton = document.getElementById('vagrantsButton');

    if (kingpinsButton) {
        kingpinsButton.addEventListener('click', function () {
            window.location.href = '/kingpins';
        });
    }

    if (vivacityButton) {
        vivacityButton.addEventListener('click', function () {
            window.location.href = '/vivacity';
        });
    }

    if (vagrantsButton) {
        vagrantsButton.addEventListener('click', function () {
            window.location.href = '/vagrants';
        });
    }
});

document.getElementById('signupButton').addEventListener('click', function () {
    window.location.href = '/auth/battlenet';
});

document.getElementById('profileButton').addEventListener('click', async function () {
    const user = await get_user();
    if (user.NickName) {
        window.location.href = `/profile/${user.UserID}`;
    } else {
        console.log(user.NickName)
        console.log(user.UserID)
        console.error("User NickName not found");
    }
});
