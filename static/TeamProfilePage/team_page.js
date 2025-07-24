// Extract teamID from URL
  const teamID = window.location.pathname.split('/').pop();

  // Check authentication status
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

  // // Fetch team data
  // fetch(`/api/teams?team_id=${teamID}`)
  //   .then(response => {
  //     if (!response.ok) throw new Error('Team not found');
  //     return response.json();
  //   })
  //   .then(data => {
  //     document.getElementById('teamName').textContent = data.name;
  //   })
  //   .catch(error => {
  //     console.error('Error fetching team:', error);
  //     document.getElementById('teamName').textContent = 'Team Not Found';
  //   });

  // // Fetch team members (now includes usernames)
  // fetch(`/api/team_members?team_id=${teamID}`)
  //   .then(response => {
  //     if (!response.ok) throw new Error('No team members found');
  //     return response.json();
  //   })
  //   .then(members => {
  //     const playersGrid = document.getElementById('playersGrid');
  //     if (!members || members.length === 0) {
  //       playersGrid.innerHTML = '<p class="text-cyan-400">No players found for this team.</p>';
  //       return;
  //     }
      
  //     // Directly map the members to create the player cards
  //     playersGrid.innerHTML = members.map(member => `
  //       <div class="player-card">
  //         <span class="text-xl text-cyan-400">${member.Username}</span>
  //       </div>
  //     `).join('');
  //   })
  //   .catch(error => {
  //     console.error('Error fetching team members:', error);
  //     document.getElementById('playersGrid').innerHTML = '<p class="text-cyan-400">Error loading players.</p>';
  //   });

  // This script runs when the team profile page is loaded.
document.addEventListener('DOMContentLoaded', () => {
    // Create a new URLSearchParams object from the current URL's query string
    const params = new URLSearchParams(window.location.search);
    // Get the value of the 'teamId' parameter
    const teamId = params.get('teamId');

    // If no teamId is found in the URL, display an error and stop.
    if (!teamId) {
        document.body.innerHTML = '<h1 style="color: white; text-align: center;">Error: No Team ID provided in URL.</h1>';
        return;
    }

    // Fetch the specific team's data from our new API endpoint
    fetch(`/api/teams/${teamId}`)
        .then(response => {
            // If the server response is not OK (e.g., 404 Not Found, 500 Server Error),
            // throw an error to be caught by the .catch() block.
            if (!response.ok) {
                return response.text().then(text => { throw new Error(text || 'Team not found') });
            }
            // Otherwise, parse the JSON data from the response.
            return response.json();
        })
        .then(data => {
            // Once we have the data, use it to populate the page.
            populateTeamData(data);
        })
        .catch(error => {
            // If any error occurred during the fetch, log it and display an error message.
            console.error('Error fetching team data:', error);
            const teamNameElement = document.getElementById('team-name');
            const membersContainer = document.getElementById('team-members');
            if (teamNameElement) {
                teamNameElement.textContent = 'Error Loading Team';
            }
            if (membersContainer) {
                membersContainer.innerHTML = `<p style="color: #ffdddd;">Could not load team details: ${error.message}</p>`;
            }
        });
});

/**
 * Populates the HTML elements with data fetched from the API.
 * @param {object} data - The team profile data, including name and members list.
 */
function populateTeamData(data) {
    // Set the team name in the <h1> tag
    const teamNameElement = document.getElementById('team-name');
    if (teamNameElement) {
        teamNameElement.textContent = data.name || 'Unnamed Team';
    }

    // Populate the list of team members
    const membersContainer = document.getElementById('team-members');
    if (membersContainer) {
        membersContainer.innerHTML = '<h2>Members</h2>'; // Reset content with header

        if (data.members && data.members.length > 0) {
            const list = document.createElement('ul');
            list.className = 'members-list'; // Add a class for styling

            data.members.forEach(member => {
                const listItem = document.createElement('li');
                listItem.className = 'member-item';
                // Display username and role for each member
                listItem.innerHTML = `<span class="username">${member.username}</span> - <span class="role">${member.role}</span>`;
                list.appendChild(listItem);
            });
            membersContainer.appendChild(list);
        } else {
            // If there are no members, display a message.
            membersContainer.innerHTML += '<p>This team has no members yet.</p>';
        }
    }
}