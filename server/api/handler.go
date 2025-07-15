package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// AvailabilityHandler handles GET and POST requests for user availability.
func AvailabilityHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
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
		} else if r.Method == "POST" {
			var req struct {
				UserID    int    `json:"user_id"`
				TeamID    int    `json:"team_id"`
				Day       string `json:"day"`
				Time      string `json:"time"`
				Available bool   `json:"available"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM team_members WHERE user_id = $1 AND team_id = $2", req.UserID, req.TeamID).Scan(&count)
			if err != nil || count == 0 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			var slotID int
			err = db.QueryRow("SELECT slot_id FROM time_slots WHERE weekday = $1 AND time = $2", req.Day, req.Time).Scan(&slotID)
			if err != nil {
				http.Error(w, "Invalid slot", http.StatusBadRequest)
				return
			}
			_, err = db.Exec(`
				INSERT INTO availability (user_id, slot_id, available)
				VALUES ($1, $2, $3)
				ON CONFLICT (user_id, slot_id) DO UPDATE SET available = $3
			`, req.UserID, slotID, req.Available)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func TeamHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Retrieve team information based on TeamID or TeamName
			var req struct {
				TeamID   int    `json:"team_id"`
				TeamName string `json:"team_name"`
			}
			// Decode the request body
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			// Validate that at least one of TeamID or TeamName is provided
			if req.TeamID == 0 && req.TeamName == "" {
				row := db.QueryRow("SELECT id, name FROM teams")
				var id int
				var name string
				if err := row.Scan(&id, &name); err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "No teams found", http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{"id": id, "name": name})
				return
			}
			if req.TeamID != 0 {
				row := db.QueryRow("SELECT name FROM teams WHERE id = $1", req.TeamID)
				var name string
				if err := row.Scan(&name); err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "Team not found", http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{"id": req.TeamID, "name": name})
			} else if req.TeamName != "" {
				row := db.QueryRow("SELECT id FROM teams WHERE name = $1", req.TeamName)
				var id int
				if err := row.Scan(&id); err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "Team not found", http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{"id": id, "name": req.TeamName})
			} else {
				// Error Handling
				http.Error(w, "Team ID or Team Name must be provided", http.StatusBadRequest)
				return
			}
		} else if r.Method == "POST" {
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
		} else if r.Method == "DELETE" {
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
		} else if r.Method == "PUT" {
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
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func PlayerHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Retrieve user username based on UserID
			var req struct {
				UserID string `json:"user_id"`
			}
			// Decode the request body
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.UserID == "" {
				http.Error(w, "User ID is required", http.StatusBadRequest)
				return
			}
			row := db.QueryRow("SELECT username FROM users WHERE user_id = $1", req.UserID)
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
		} else if r.Method == "DELETE" {
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
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func TeamMembersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			var req struct {
				TeamID int `json:"team_id"`
			}
			// Decode the request body
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid input", http.StatusBadRequest)
				return
			}
			if req.TeamID == 0 {
				http.Error(w, "Team ID is required", http.StatusBadRequest)
				return
			}
			rows, err := db.Query("SELECT user_id FROM team_members WHERE team_id = $1", req.TeamID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			var members []int
			for rows.Next() {
				var userID int
				err := rows.Scan(&userID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				members = append(members, userID)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(members)
		} else if r.Method == "POST" {
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
			// Check if the user is already a member of the team
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM team_members WHERE user_id = $1 AND team_id = $2", req.UserID, req.TeamID).Scan(&count)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if count > 0 {
				http.Error(w, "User is already a member of the team", http.StatusBadRequest)
				return
			}
			// Insert the user into the team_members table
			_, err = db.Exec("INSERT INTO team_members (user_id, team_id) VALUES ($1, $2)", req.UserID, req.TeamID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
		} else if r.Method == "DELETE" {
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
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
func TimeSlotsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
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
		} else if r.Method == "POST" {
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
		} else if r.Method == "DELETE" {
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
		} else if r.Method == "PUT" {
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
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
