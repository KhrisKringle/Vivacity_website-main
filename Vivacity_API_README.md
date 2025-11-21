# Vivacity Website – Schedule Manager API

## Overview

**Vivacity** is a team-based scheduling application inspired by When2Meet, designed for esports organizations to manage teams, players, events, and availability efficiently.

- **Backend:** Go (Golang), PostgreSQL
- **Frontend:** Server-side rendering (SSR, in progress)
- **Deployment:** Dockerized

---

## API Overview

All API endpoints are prefixed with `/api`.

---

## Team Endpoints

### `POST /api/teams`

**Description:** Create a new team.

**Request Body:**
```json
{
  "name": "Alpha Squad"
}
```

**Response:**
- `201 Created`
- Returns created team ID

**Example:**
```bash
curl -X POST http://localhost:8080/api/teams -d '{"name":"Alpha Squad"}'
```

---

### `GET /api/teams`

**Description:** List all teams.

**Response:**
```json
[
  {"name": "Alpha Squad"},
  {"name": "Beta Crew"}
]
```

**Example:**
```bash
curl http://localhost:8080/api/teams
```

---

### `GET /api/teams/{team_id}`

**Description:** Get a specific team's info including players.

**Response:**
```json
{
  "name": "Alpha Squad",
  "players": [
    {
      "player_id": 1,
      "name": "John",
      "battletag": "John#1234",
      "position": "Tank"
    },
    {
      "player_id": 2,
      "name": Doe",
      "battletag": "Doe#5678",
      "position": "Support"
    }
  ]
}
```

**Example:**
```bash
curl http://localhost:8080/api/teams/1
```

---

### `DELETE /api/teams/{team_id}`

**Description:** Delete a team.

**Response:** `204 No Content`

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/teams/1
```

---

### `GET /api/teams/{team_id}/events`

**Description:** Get all events for a team.

**Response:**
```json
[
  {
    "event_id": 1,
    "title": "Team Meeting",
    "start_time": "2025-05-01T19:00:00Z",
    "end_time": "2025-05-01T21:00:00Z"
  }
]
```

**Example:**
```bash
curl http://localhost:8080/api/teams/1/events
```

---

## Player Endpoints

### `POST /api/players`

**Description:** Create a player and assign to a team. PROBABLY WRONG

**Request Body:**
```json
{
  "team_id": 1,
  "name": "John",
  "battletag": "John#1234",
  "position": "Leader"
}
```

**Response:**
- `201 Created`
- Returns player ID

**Example:**
```bash
curl -X POST http://localhost:8080/api/players -d '{"team_id":1,"name":"John","battletag":"John#1234","position":"Leader"}'
```

---

### `GET /api/players`

**Description:** List all players. Probably will deperacate

**Response:**
```json
[
  {
    "player_id": 1,
    "name": "John",
    "battletag": "John#1234",
    "position": "Leader",
    "team_id": 1
  }
]
```

**Example:**
```bash
curl http://localhost:8080/api/players
```

---

### `GET /api/players/{player_id}`

**Description:** Get a specific player's info.

**Response:**
```json
{
  "player_id": 1,
  "team_id": 1,
  "name": "John",
  "battletag": "John#1234",
  "position": "Tank"
}
```

**Example:**
```bash
curl http://localhost:8080/api/players/1
```

---

### `PUT /api/players/{player_id}`

**Description:** Update a player's info.

**Request Body:**
```json
{
  "team_id": 1,
  "name": "Johnny",
  "battletag": "Johnny#5678",
  "position": "DPS"
}
```

**Response:** `200 OK`

**Example:**
```bash
curl -X PUT http://localhost:8080/api/players/1 \
     -H "Content-Type: application/json" \
     -d '{"team_id":1,"name":"Johnny","battletag":"Johnny#5678","position":"Member"}'
```

---

## Notes

- Future features: Discord integration, email reminders.
- Deployment assumes use of Docker
