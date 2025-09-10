// Check authentication status and set button states
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