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

                    // --- MODIFICATION IS HERE ---
                    // Combine the day and time into one readable string.
                    const formattedDateTime = `${item.time}`;

                    // Update the HTML to show the activity on the left
                    // and the new formatted date/time on the right.
                    scheduleItem.innerHTML = `
                        <div>
                            <p class="activity">${item.weekday}</p>
                        </div>
                        <p class="time">${formattedDateTime}</p>
                    `;

                    // *** ADD THIS EVENT LISTENER ***
                    // This will toggle the 'selected' class on click.
                    scheduleItem.addEventListener('click', () => {
                        scheduleItem.classList.toggle('selected');
                    });

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