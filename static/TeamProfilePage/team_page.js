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

  // Fetch team data
  fetch(`/api/teams?team_id=${teamID}`)
    .then(response => {
      if (!response.ok) throw new Error('Team not found');
      return response.json();
    })
    .then(data => {
      document.getElementById('teamName').textContent = data.name;
    })
    .catch(error => {
      console.error('Error fetching team:', error);
      document.getElementById('teamName').textContent = 'Team Not Found';
    });

  // Fetch team members (now includes usernames)
  fetch(`/api/team_members?team_id=${teamID}`)
    .then(response => {
      if (!response.ok) throw new Error('No team members found');
      return response.json();
    })
    .then(members => {
      const playersGrid = document.getElementById('playersGrid');
      if (!members || members.length === 0) {
        playersGrid.innerHTML = '<p class="text-cyan-400">No players found for this team.</p>';
        return;
      }
      
      // Directly map the members to create the player cards
      playersGrid.innerHTML = members.map(member => `
        <div class="player-card">
          <span class="text-xl text-cyan-400">${member.Username}</span>
        </div>
      `).join('');
    })
    .catch(error => {
      console.error('Error fetching team members:', error);
      document.getElementById('playersGrid').innerHTML = '<p class="text-cyan-400">Error loading players.</p>';
    });