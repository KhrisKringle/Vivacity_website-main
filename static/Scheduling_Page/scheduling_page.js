document.addEventListener('DOMContentLoaded', () => {
    const params = new URLSearchParams(window.location.search);
    const team_id = params.get('team_id');
    const scheduleContainer = document.getElementById('scheduleContainer');
    const submitButton = document.getElementById('submitAvailability');

    // This array will store the selected timeslots
    let selectedTimeslots = [];

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
                            <p class="weekday">${item.weekday}</p>
                        </div>
                        <p class="time">${formattedDateTime}</p>
                    `;

                    // *** ADD THIS EVENT LISTENER ***
                    // This will toggle the 'selected' class on click.
                    scheduleItem.addEventListener('click', () => {
                        scheduleItem.classList.toggle('selected');

                        const timeSlot = {
                            weekday: item.weekday,
                            time: item.time
                        };
                        // Check if the timeslot is already selected
                        const index = selectedTimeslots.findIndex(slot => slot.day === timeSlot.day && slot.time === timeSlot.time);

                        if (index > -1) {
                            // If it is, remove it from the array
                            selectedTimeslots.splice(index, 1);
                        } else {
                            // If it's not, add it to the array
                            selectedTimeslots.push(timeSlot);
                        }
                        
                        console.log('Selected Times:', selectedTimeslots); // For debugging
                    
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
        // --- NEW: Event listener for the submit button ---
    submitButton.addEventListener('click', () => {
        if (selectedTimeslots.length === 0) {
            alert('Please select at least one timeslot.');
            return;
        }

        // POST the selected times to a new API endpoint
        fetch(`/api/teams/${team_id}/availability`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                // You might want to include a player ID here in a real application
                player_id: 'current_player_id', 
                selected_slots: selectedTimeslots
            })
        })
        .then(response => {
            if (!response.ok) throw new Error('Failed to submit availability.');
            return response.json();
        })
        .then(data => {
            alert('Availability submitted successfully!');
            console.log(data);
            // Optionally, redirect or clear selections
            selectedTimeslots = [];
            document.querySelectorAll('.schedule-item.selected').forEach(card => {
                card.classList.remove('selected');
            });
        })
        .catch(error => {
            console.error('Error submitting availability:', error);
            alert(error.message);
        });
    });
});