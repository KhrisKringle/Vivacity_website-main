document.addEventListener('DOMContentLoaded', function () {
    const vivacitySchedualButton = document.getElementById('kingpinsSchedualButton');
    const backToVivacityButton = document.getElementById('backToKingpinsButton');

    if (vivacitySchedualButton) {
        vivacitySchedualButton.addEventListener('click', function () {
            window.location.href = '/kingpins/schedual';
        });
    }

    if (backToVivacityButton) {
        backToVivacityButton.addEventListener('click', function () {
            window.location.href = '/kingpins';
        });
    }
});
