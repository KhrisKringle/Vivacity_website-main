document.addEventListener('DOMContentLoaded', function () {
    const vagrantsSchedualButton = document.getElementById('vagrantsSchedualButton');
    const backToVagrantsButton = document.getElementById('backToVagrantsButton');

    if (vagrantsSchedualButton) {
        vagrantsSchedualButton.addEventListener('click', function () {
            window.location.href = '/vagrants/schedual';
        });
    }

    if (backToVagrantsButton) {
        backToVagrantsButton.addEventListener('click', function () {
            window.location.href = '/vagrants';
        });
    }
});