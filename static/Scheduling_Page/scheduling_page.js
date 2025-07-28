document.addEventListener('DOMContentLoaded', () => {
    const params = new URLSearchParams(window.location.search);
    const team_id = params.get('team_id');

    if (!team_id) {
        document.getElementById('teamName').textContent = 'Error: No Team ID Provided';
        return;
    }

    // Fetch team name first
    fetch(`/api/teams/${team_id}`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Team not found');
            }
            return response.json();
        })
        .then(teamData => {
            document.getElementById('teamName').textContent = `${teamData.name} Schedule`;
        })
        .catch(error => {
            console.error('Error fetching team name:', error);
            document.getElementById('teamName').textContent = 'Schedule Not Found';
        });

    // Fetch schedule data
    fetch(`/api/teams/${team_id}/schedule`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Could not load schedule');
            }
            return response.json();
        })
        .then(schedule => {
            const scheduleContainer = document.getElementById('scheduleContainer');
            scheduleContainer.innerHTML = ''; // Clear loading text

            if (schedule && schedule.length > 0) {
                schedule.forEach(item => {
                    const scheduleItem = document.createElement('div');
                    scheduleItem.className = 'schedule-item';

                    scheduleItem.innerHTML = `
                        <div>
                            <p class="day">${item.day}</p>
                            <p class="activity">${item.activity}</p>
                        </div>
                        <p class="time">${item.time}</p>
                    `;
                    scheduleContainer.appendChild(scheduleItem);
                });
            } else {
                scheduleContainer.innerHTML = '<p>No schedule has been set for this team yet.</p>';
            }
        })
        .catch(error => {
            console.error('Error fetching schedule:', error);
            const scheduleContainer = document.getElementById('scheduleContainer');
            scheduleContainer.innerHTML = `<p>${error.message}</p>`;
        });
});