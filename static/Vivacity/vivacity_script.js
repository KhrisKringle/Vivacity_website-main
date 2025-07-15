document.addEventListener('DOMContentLoaded', function () {
    const vivacitySchedualButton = document.getElementById('vivacitySchedualButton');
    const backToVivacityButton = document.getElementById('backToVivacityButton');

    if (vivacitySchedualButton) {
        vivacitySchedualButton.addEventListener('click', function () {
            window.location.href = '/vivacity/schedual';
        });
    }

    if (backToVivacityButton) {
        backToVivacityButton.addEventListener('click', function () {
            window.location.href = '/vivacity';
        });
    }
});
