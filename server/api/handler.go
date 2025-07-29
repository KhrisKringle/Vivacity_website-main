package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type AvailabilitySlot struct {
	Day  string `json:"day"`
	Time string `json:"time"`
}

// Represents the entire JSON object from the frontend
type AvailabilityRequest struct {
	SelectedSlots []AvailabilitySlot `json:"selected_slots"`
}

// AvailabilityHandler handles GET and POST requests for user availability.
func AvailabilityHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userID := r.URL.Query().Get("user_id")
			teamID := r.URL.Query().Get("team_id")
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM team_members WHERE user_id = $1 AND team_id = $2", userID, teamID).Scan(&count)
			if err != nil || count == 0 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			rows, err := db.Query(`
				SELECT t.weekday, t.time, a.available
				FROM time_slots t
				LEFT JOIN availability a ON t.slot_id = a.slot_id AND a.user_id = $1
			`, userID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			var slots []map[string]any
			for rows.Next() {
				var day, time string
				var available sql.NullBool
				err := rows.Scan(&day, &time, &available)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				slots = append(slots, map[string]any{
					"day":       day,
					"time":      time,
					"available": available.Valid && available.Bool,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(slots)
		case http.MethodPost:
			var req struct {
				AvailabilityRequest
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if len(req.SelectedSlots) == 0 {
				http.Error(w, "User ID and selected slots are required", http.StatusBadRequest)
				return
			}
			// Check if the user is a member of the team
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM team_members WHERE user_id = $1 AND team_id = $2", r.Context().Value("user_id"), r.Context().Value("team_id")).Scan(&count)
			if err != nil || count == 0 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Clear existing availability for the user
			_, err = db.Exec("DELETE FROM availability WHERE user_id = $1", r.Context().Value("user_id"))
			if err != nil {
				http.Error(w, "Failed to clear existing availability", http.StatusInternalServerError)
				return
			}

			// Insert new availability slots
			for _, slot := range req.SelectedSlots {
				_, err = db.Exec("INSERT INTO availability (user_id, slot_id) VALUES ($1, (SELECT id FROM time_slots WHERE weekday = $2 AND time = $3 LIMIT 1))",
					r.Context().Value("user_id"), slot.Day, slot.Time)
				if err != nil {
					http.Error(w, "Failed to insert new availability slot", http.StatusInternalServerError)
					return
				}
			}

			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func TeamHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Extract the team ID from the URL path, e.g., "/api/teams/1"
			teamIDStr := strings.TrimPrefix(r.URL.Path, "/api/teams/")
			teamID, err := strconv.Atoi(teamIDStr)
			if err != nil {
				http.Error(w, "Invalid team ID provided in URL", http.StatusBadRequest)
				return
			}

			// --- Step 1: Fetch the team's basic details ---
			var team Team
			err = db.QueryRow("SELECT id, name FROM teams WHERE id = $1", teamID).Scan(&team.ID, &team.Name)
			if err != nil {
				// If no team is found, return a 404
				if err == sql.ErrNoRows {
					http.Error(w, "Team not found", http.StatusNotFound)
					return
				}
				// For any other database error, return a 500
				http.Error(w, "Database error fetching team", http.StatusInternalServerError)
				return
			}

			// --- Step 2: Fetch the list of members for that team ---
			rows, err := db.Query(`
            SELECT u.id, u.username, tm.role
            FROM users u
            JOIN team_members tm ON u.id = tm.user_id
            WHERE tm.team_id = $1`, teamID)
			if err != nil {
				http.Error(w, "Database error fetching team members", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var members []TeamMember
			for rows.Next() {
				var member TeamMember
				if err := rows.Scan(&member.ID, &member.Username, &member.Role); err != nil {
					http.Error(w, "Error scanning team member data", http.StatusInternalServerError)
					return
				}
				members = append(members, member)
			}

			// --- Step 3: Combine team details and members into a single response ---
			fullTeamProfile := struct {
				Team
				Members []TeamMember `json:"members"`
			}{
				Team:    team,
				Members: members,
			}

			// --- Step 4: Send the JSON response ---
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(fullTeamProfile)
		case http.MethodPost:
			// Create a new team
			var req struct {
				Name string `json:"name"`
			}
			// Decode the request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.Name == "" {
				// Error Handling
				http.Error(w, "Team Name must be provided", http.StatusBadRequest)
				return
			}
			// Insert the new team into the database
			_, err := db.Exec("INSERT INTO teams (name) VALUES ($1)", req.Name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			var req struct {
				TeamID int `json:"team_id"`
			}

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.TeamID == 0 {
				http.Error(w, "Team ID must be provided", http.StatusBadRequest)
				return
			}
			_, err := db.Exec("DELETE FROM teams WHERE id = $1", req.TeamID)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Team not found", http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPut:
			// Update an existing team
			var req struct {
				TeamID   int    `json:"team_id"`
				TeamName string `json:"team_name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.TeamName == "" {
				// Error Handling
				http.Error(w, "Team Name must be provided", http.StatusBadRequest)
				return
			}
			_, err := db.Exec("UPDATE teams SET name = $1 WHERE id = $2", req.TeamName, req.TeamID)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Team not found", http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func PlayerHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Retrieve user username based on UserID
			userIDStr := r.URL.Query().Get("user_id")
			if userIDStr == "" {
				http.Error(w, "User ID is required", http.StatusBadRequest)
				return
			}

			// Convert the user ID from string to integer
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				http.Error(w, "Invalid User ID", http.StatusBadRequest)
				return
			}
			row := db.QueryRow("SELECT username FROM users WHERE user_id = $1", userID)
			var username string
			if err := row.Scan(&username); err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "User not found", http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"username": username})
		case http.MethodDelete:
			var req struct {
				UserID int `json:"user_id"`
			}
			// Decode the request body
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.UserID == 0 {
				http.Error(w, "User ID is required", http.StatusBadRequest)
				return
			}
			// Check if the user is a member of a team
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM team_members WHERE user_id = $1", req.UserID).Scan(&count)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if count == 0 {
				http.Error(w, "User is not a member of the team", http.StatusNotFound)
				return
			}
			// Delete the user from the team_members table
			_, err = db.Exec("DELETE FROM team_members WHERE user_id = $1", req.UserID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// Delete the user from the users table
			_, err = db.Exec("DELETE FROM users WHERE user_id = $1", req.UserID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func TeamMembersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			teamIDStr := r.URL.Query().Get("team_id")
			if teamIDStr == "" {
				http.Error(w, "Team ID is required", http.StatusBadRequest)
				return
			}

			// Convert the team ID from string to integer
			teamID, err := strconv.Atoi(teamIDStr)
			if err != nil {
				http.Error(w, "Invalid Team ID", http.StatusBadRequest)
				return
			}

			// Grab the team members from the database
			rows, err := db.Query(`
				SELECT u.user_id, u.username, tm.role 
				FROM users u
				JOIN team_members tm ON u.id = tm.user_id
				WHERE tm.team_id = $1`, teamID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			var members []map[string]any
			for rows.Next() {
				var userID int
				var username string
				var role string
				err := rows.Scan(&userID, &username, &role)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				members = append(members, map[string]any{
					"user_id":  userID,
					"username": username,
					"role":     role,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(members)
		case http.MethodPost:
			var req struct {
				UserID int    `json:"user_id"`
				TeamID int    `json:"team_id"`
				Role   string `json:"role"`
			}
			// Decode the request body
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.UserID == 0 || req.TeamID == 0 {
				http.Error(w, "User ID and Team ID are required", http.StatusBadRequest)
				return
			}
			// Check if the user is already a member of the team
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM team_members WHERE user_id = $1 AND team_id = $2", req.UserID, req.TeamID).Scan(&count)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if count > 0 {
				http.Error(w, "User is already a member of this team", http.StatusBadRequest)
				return
			}
			// Insert the user into the team_members table
			_, err = db.Exec("INSERT INTO team_members (user_id, team_id, role) VALUES ($1, $2, $3)", req.UserID, req.TeamID, req.Role)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)

		case http.MethodDelete:
			var req struct {
				UserID int `json:"user_id"`
				TeamID int `json:"team_id"`
			}
			// Decode the request body
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.UserID == 0 || req.TeamID == 0 {
				http.Error(w, "User ID and Team ID are required", http.StatusBadRequest)
				return
			}
			// Delete the user from the team_members table
			_, err := db.Exec("DELETE FROM team_members WHERE user_id = $1 AND team_id = $2", req.UserID, req.TeamID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPut:
			var req struct {
				UserID int    `json:"user_id"`
				TeamID int    `json:"team_id"`
				Role   string `json:"role"`
			}
			// Decode the request body
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.UserID == 0 || req.TeamID == 0 {
				http.Error(w, "User ID and Team ID are required", http.StatusBadRequest)
				return
			}
			// Update the user's role in the team_members table
			_, err := db.Exec("UPDATE team_members SET role = $1 WHERE user_id = $2 AND team_id = $3", req.Role, req.UserID, req.TeamID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
func TimeSlotsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			var req struct {
				TeamID int `json:"team_id"`
			}
			rows, err := db.Query("SELECT weekday, time FROM time_slots WHERE team_id = $1", req.TeamID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			var slots []map[string]any
			for rows.Next() {
				var day, time string
				err := rows.Scan(&day, &time)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				slots = append(slots, map[string]any{
					"day":  day,
					"time": time,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(slots)
		case http.MethodPost:
			var req struct {
				Day    string `json:"day"`
				Time   string `json:"time"`
				TeamID int    `json:"team_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.Day == "" || req.Time == "" {
				http.Error(w, "Day and Time are required", http.StatusBadRequest)
				return
			}
			_, err := db.Exec("INSERT INTO time_slots (weekday, time, team_id) VALUES ($1, $2, $3)", req.Day, req.Time, req.TeamID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
		case http.MethodDelete:
			var req struct {
				Day    string `json:"day"`
				Time   string `json:"time"`
				TeamID int    `json:"team_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.Day == "" || req.Time == "" {
				http.Error(w, "Day and Time are required", http.StatusBadRequest)
				return
			}
			_, err := db.Exec("DELETE FROM time_slots WHERE weekday = $1 AND time = $2 AND team_id = $3", req.Day, req.Time, req.TeamID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPut:
			var req struct {
				Day     string `json:"day"`
				Time    string `json:"time"`
				TeamID  int    `json:"team_id"`
				NewDay  string `json:"new_day"`
				NewTime string `json:"new_time"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.Day == "" || req.Time == "" || req.NewDay == "" || req.NewTime == "" {
				http.Error(w, "Day, Time, New Day and New Time are required", http.StatusBadRequest)
				return
			}
			_, err := db.Exec("UPDATE time_slots SET weekday = $1, time = $2 WHERE weekday = $3 AND time = $4 AND team_id = $5", req.NewDay, req.NewTime, req.Day, req.Time, req.TeamID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func ScheduleHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Fetch all time slots
			rows, err := db.Query("SELECT slot_id, weekday, time FROM time_slots")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			var slots []map[string]any
			for rows.Next() {
				var slotID int
				var weekday, time string
				err := rows.Scan(&slotID, &weekday, &time)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				slots = append(slots, map[string]any{
					"slot_id": slotID,
					"weekday": weekday,
					"time":    time,
				})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(slots)

		case http.MethodPost:
			// Create a new time slot
			var req struct {
				Weekday string `json:"weekday"`
				Time    string `json:"time"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.Weekday == "" || req.Time == "" {
				http.Error(w, "Weekday and Time are required", http.StatusBadRequest)
				return
			}
			_, err := db.Exec("INSERT INTO time_slots (weekday, time) VALUES ($1, $2)", req.Weekday, req.Time)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)

		case http.MethodDelete:
			// Delete a time slot
			var req struct {
				SlotID int `json:"slot_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.SlotID == 0 {
				http.Error(w, "Slot ID is required", http.StatusBadRequest)
				return
			}
			_, err := db.Exec("DELETE FROM time_slots WHERE slot_id = $1", req.SlotID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodPut:
			// Update a time slot
			var req struct {
				SlotID  int    `json:"slot_id"`
				Weekday string `json:"weekday"`
				Time    string `json:"time"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.SlotID == 0 || req.Weekday == "" || req.Time == "" {
				http.Error(w, "Slot ID, Weekday and Time are required", http.StatusBadRequest)
				return
			}
			_, err := db.Exec("UPDATE time_slots SET weekday = $1, time = $2 WHERE slot_id = $3", req.Weekday, req.Time, req.SlotID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
