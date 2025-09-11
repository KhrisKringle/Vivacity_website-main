// This script runs once the entire HTML document has been loaded and parsed.
document.addEventListener('DOMContentLoaded', () => {
    fetch('/auth/status')
      .then(response => {
        if (response.status === 401) {
          document.getElementById('profileButton').classList.add('hidden');
          document.getElementById('authButton').textContent = 'Login';
          document.getElementById('authButton').href = '/auth/battlenet';
        } else {
          return response.json().then(user => {
            document.getElementById('profileButton').href = `/profile/${user.UserID}`;
            document.getElementById('profileButton').classList.remove('hidden');
            document.getElementById('authButton').textContent = 'Logout';
            document.getElementById('authButton').href = '/logout';
          });
        }
      })
      .catch(error => {
        console.error('Error checking auth:', error);
        document.getElementById('profileButton').classList.add('hidden');
        document.getElementById('authButton').textContent = 'Login';
        document.getElementById('authButton').href = '/auth/battlenet';
      });
    // --- Step 1: Fetch All Teams from the API ---
    // This function call retrieves the list of all available teams from the backend.
    fetch('/api/teams')
        .then(response => {
            // If the server returns an error (e.g., 500 Internal Server Error), handle it.
            if (!response.ok) {
                throw new Error('Failed to fetch teams. Please try again later.');
            }
            // If the response is successful, parse the JSON data.
            return response.json();
        })
        .then(teams => {
            // --- Step 2: Populate the Page with Fetched Teams ---
            // Get the container where the team cards will be displayed.
            const teamsGrid = document.getElementById('teamsGrid');
            // Clear any "loading..." text.
            teamsGrid.innerHTML = '';

            // If there are teams, loop through them and create a card for each.
            if (teams && teams.length > 0) {
                teams.forEach(team => {
                    const teamCard = document.createElement('a');
                    // *** Correction is here ***
                    // The link now correctly points to the team profile page,
                    // passing the team's ID in the URL.
                    teamCard.href = `/team-profile?team_id=${team.ID}`;
                    teamCard.className = 'team-card'; // Use this class for styling in your CSS.

                    teamCard.innerHTML = `
                        <h3 class="text-xl text-cyan-400">${team.name}</h3>
                        <p class="text-gray-400">${team.game || 'No game specified'}</p>
                    `;
                    teamsGrid.appendChild(teamCard);
                });
            } else {
                // If there are no teams, display a message.
                teamsGrid.innerHTML = '<p class="col-span-full text-center text-gray-400">No teams have been created yet.</p>';
            }
        })
        .catch(error => {
            // If any part of the fetch process fails, display an error message.
            console.error('Error fetching teams:', error);
            const teamsGrid = document.getElementById('teamsGrid');
            teamsGrid.innerHTML = `<p class="text-red-400 col-span-full text-center">${error.message}</p>`;
        });
});