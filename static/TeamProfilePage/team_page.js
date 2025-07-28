// This script runs once the entire HTML document has been loaded and parsed.
document.addEventListener('DOMContentLoaded', () => {
    
    // --- Step 1: Get the Team ID from the URL ---
    // The URL looks like "/team-profile?team_id=1". This code extracts the "1".
    const params = new URLSearchParams(window.location.search);
    const team_id = params.get('team_id');

    // If no team_id is found in the URL, we can't load a profile. 
    // Display an error message and stop the script.
    if (!team_id) {
        document.getElementById('team_name').textContent = 'Error: No Team ID Provided';
        document.getElementById('playersGrid').innerHTML = '<p class="text-red-400">Please go back to the teams page and select a team.</p>';
        console.error("No team_id found in URL parameters.");
        return; // Stop the function from proceeding further
    }

    // --- Step 2: Fetch the Team's Data from the API ---
    // Use the team_id to call the specific API endpoint for that team.
    fetch(`/api/teams/${team_id}`)
        .then(async response => {
            // If the server returns an error (like 404 Not Found), handle it.
            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || `Team with ID ${team_id} not found`);
            }
            // If the response is successful, parse the JSON data.
            return response.json();
        })
        .then(data => {
            // --- Step 3: Populate the Page with the Fetched Data ---
            // Call our helper function to update the HTML with the team's info.
            populateTeamData(data);
        })
        .catch(error => {
            // If any part of the fetch process fails, display an informative error message.
            console.error('Error fetching team data:', error);
            const teamNameElement = document.getElementById('team_name');
            if (teamNameElement) {
                teamNameElement.textContent = 'Error: Team Could Not Be Loaded';
            }
            const playersGrid = document.getElementById('playersGrid');
            if (playersGrid) {
                playersGrid.innerHTML = `<p style="color: #ffdddd;">Details: ${error.message}</p>`;
            }
        });

    // --- Step 4: Check Authentication Status ---
    // This part runs independently to update the navigation buttons.
    fetch('/auth/user')
        .then(response => {
            if (response.status === 401) {
                // User is not logged in
                document.getElementById('profileButton').classList.add('hidden');
                document.getElementById('authButton').textContent = 'Login';
                document.getElementById('authButton').href = '/login';
            } else {
                // User is logged in
                return response.json().then(user => {
                    document.getElementById('profileButton').href = `/profile/${user.UserID}`;
                    document.getElementById('profileButton').classList.remove('hidden');
                    document.getElementById('authButton').textContent = 'Logout';
                    document.getElementById('authButton').href = '/logout/battlenet';
                });
            }
        })
        .catch(error => {
            // Handle any errors during the auth check
            console.error('Error checking auth status:', error);
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
    const teamNameElement = document.getElementById('team_name');
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
