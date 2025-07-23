# Vivacity_website

## Schedule Manager API and Frontend Documentation
This document outlines the backend API endpoints and frontend requirements for the Schedule Manager, a team-based scheduling application similar to When2Meet. The app allows organizations to manage teams, players, events, and availability, with a focus on scheduling team activities. The backend is built with Go and PostgreSQL, running in a Dockerized environment, while the frontend is under development.
## Backend API
The backend provides RESTful endpoints for managing Teams, Players, and Events, with planned integrations for Discord and potential email notifications. All endpoints are prefixed with /api.
## Teams
Endpoints for creating, managing, and retrieving team information.

### POST /api/teams
  Description: Create a new team.
<pre style='font-size: 1.25rem; line-height: 1.25;'>
  Request Body:{
    "name": "string" // Team name (required)
  }
</pre>

  - Response: HTTP 201 Created, team ID in body.
  - Example:curl -X POST http://localhost:8080/api/teams -d '{"name":"Alpha Squad"}'


### GET /api/teams

Description: List all teams in the organization.
<pre style='font-size: 1.25rem; line-height: 1.25;'>
  Response:[
    {"name": "Alpha Squad"},
    {"name": "Beta Crew"}
  ]
</pre>

  - Example:curl http://localhost:8080/api/teams




### GET /api/teams/{team_id}
  Description: Get details for a specific team, including its players.
  Parameters:
  team_id: Team identifier (path parameter).

<pre style='font-size: 1.25rem; line-height: 1.25;'>
  Response:{
    "name": "Alpha Squad",
    "players": [
      {"player_id": 1, "name": "John", "battletag": "John#1234", "position": "Leader"}
    ]
  }
</pre>

  - Example:curl http://localhost:8080/api/teams/1

### DELETE /api/teams/{team_id}
  Description: Delete a team.
  Parameters:
  team_id: Team identifier (path parameter).


  - Response: HTTP 204 No Content.
  - Example:curl -X DELETE http://localhost:8080/api/teams/1




### GET /api/teams/{team_id}/events

Description: Get all events associated with a team.
Parameters:
team_id: Team identifier (path parameter).

<pre style='font-size: 1.25rem; line-height: 1.25;'>
  Response:[
    {
      "event_id": 1,
      "title": "Team Meeting",
      "start_time": "2025-05-01T19:00:00Z",
      "end_time": "2025-05-01T21:00:00Z"
    }
  ]
</pre>

  - Example:curl http://localhost:8080/api/teams/1/events





## Players
  Endpoints for managing players, their details, and availability.

### POST /api/players

Description: Create a new player and assign them to a team.
<pre style='font-size: 1.25rem; line-height: 1.25;'>
  Request Body:{
    "team_id": 1,
    "name": "string",
    "battletag": "string", // e.g., "John#1234"
    "position": "string"   // e.g., "Leader", "Member"
  }
</pre>

  - Response: HTTP 201 Created, player ID in body.
  - Example:curl -X POST http://localhost:8080/api/players -d '{"team_id":1,"name":"John","battletag":"John#1234","position":"Leader"}'




### GET /api/players
Description: List all players in the organization.
<pre style='font-size: 1.25rem; line-height: 1.25;'>
  Response:[
    {"player_id": 1, "name": "John", "battletag": "John#1234", "position": "Leader", "team_id": 1}
  ]
</pre>
  - Example:curl http://localhost:8080/api/players

### GET /api/players/{player_id}

Description: Get details for a specific player.
<pre style='font-size: 1.25rem; line-height: 1.25;'>
  Parameters:
  player_id: Player identifier (path parameter).


  Response:{
    "player_id": 1,
    "team_id": 1,
    "name": "John",
    "battletag": "John#1234",
    "position": "Leader"
  }
</pre>

  - Example:curl http://localhost:8080/api/players/1




PUT /api/players/{player_id}

Description: Update a player’s details.
Parameters:
player_id: Player identifier (path parameter).


Request Body:



