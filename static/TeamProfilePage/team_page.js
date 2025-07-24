// This script runs when the team profile page (team_page.html) is loaded.
document.addEventListener('DOMContentLoaded', () => {
    // --- Step 1: Get the Team ID from the URL ---
    // The URL looks like "/team-profile?teamId=1". This code extracts the "1".
    const params = new URLSearchParams(window.location.search);
    const teamId = params.get('teamId');

    // If no teamId is found, we can't load a profile. Display an error and stop.
    if (!teamId) {
        document.body.innerHTML = '<h1 style="color: white; text-align: center;">Error: No Team ID was provided in the URL.</h1>';
        return;
    }

    // --- Step 2: Fetch the Team's Data from the API ---
    // Use the teamId to call the specific API endpoint for that team.
    fetch(`/api/teams/${teamId}`)
        .then(response => {
            // If the server returns an error (like 404 Not Found), handle it.
            if (!response.ok) {
                return response.text().then(text => { throw new Error(text || 'Team not found') });
            }
            // If the response is successful, parse the JSON data.
            return response.json();
        })
        .then(data => {
            // --- Step 3: Populate the Page with the Fetched Data ---
            // Call our helper function to update the HTML.
            populateTeamData(data);
        })
        .catch(error => {
            // If any part of the fetch process fails, display an error message.
            console.error('Error fetching team data:', error);
            const teamNameElement = document.getElementById('teamName');
            if (teamNameElement) {
                teamNameElement.textContent = 'Error: Team Could Not Be Loaded';
            }
            const playersGrid = document.getElementById('playersGrid');
            if (playersGrid) {
                playersGrid.innerHTML = `<p style="color: #ffdddd;">Details: ${error.message}</p>`;
            }
        });

    // --- Step 4: Check Authentication Status (Your existing code, works perfectly) ---
    fetch('/auth/user')
        .then(response => {
            if (response.status === 401) {
                document.getElementById('profileButton').classList.add('hidden');
                document.getElementById('authButton').textContent = 'Login';
                document.getElementById('authButton').href = '/login';
            } else {
                return response.json().then(user => {
                    document.getElementById('profileButton').href = `/profile/${user.UserID}`;
                    document.getElementById('profileButton').classList.remove('hidden');
                    document.getElementById('authButton').textContent = 'Logout';
                    document.getElementById('authButton').href = '/logout/battlenet';
                });
            }
        })
        .catch(error => {
            console.error('Error checking auth:', error);
            document.getElementById('profileButton').classList.add('hidden');
            document.getElementById('authButton').textContent = 'Login';
            document.getElementById('authButton').href = '/login';
        });
});

/**
 * Populates the HTML elements with data fetched from the API.
 * @param {object} data - The team profile data, including name and members list.
 */
function populateTeamData(data) {
    // Set the team name in the <h2> tag
    const teamNameElement = document.getElementById('teamName');
    if (teamNameElement) {
        teamNameElement.textContent = data.name || 'Unnamed Team';
    }

    // Get the grid container for the players.
    const playersGrid = document.getElementById('playersGrid');
    if (playersGrid) {
        // Clear any "loading..." text.
        playersGrid.innerHTML = '';

        if (data.members && data.members.length > 0) {
            // If the team has members, loop through them and create a card for each.
            data.members.forEach(member => {
                const playerCard = document.createElement('div');
                playerCard.className = 'player-card'; // Use this class for styling in your CSS.
                
                playerCard.innerHTML = `
                    <span class="text-xl text-cyan-400">${member.username}</span>
                    <span class="text-lg text-orange-500">${member.role}</span>
                `;
                playersGrid.appendChild(playerCard);
            });
        } else {
            // If there are no members, display a message.
            playersGrid.innerHTML = '<p class="text-cyan-400 col-span-full">This team has no players yet.</p>';
        }
    }
}
