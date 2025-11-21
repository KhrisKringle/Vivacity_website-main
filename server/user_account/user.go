package user_account

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
)

type User struct {
	Username string
	Team     string
}

// and returns their internal database ID and their team ID.
func HandleBlizzardAuth(db *sql.DB, user goth.User) (err error) {
	// Check if the user already exists in the database by their Blizzard UserID
	blizzardUserID, _ := strconv.ParseInt(user.UserID, 10, 64)
	query := "SELECT id FROM users WHERE user_id = $1" // Assuming 'id' is your primary key
	err = db.QueryRow(query, blizzardUserID).Scan(&user.UserID)

	if err == sql.ErrNoRows {
		// If the user does not exist, insert a new record and get their new primary key ID
		insertQuery := "INSERT INTO users (username, user_id) VALUES ($1, $2) RETURNING id"
		err = db.QueryRow(insertQuery, user.NickName, blizzardUserID).Scan(&user.UserID)
		if err != nil {
			return fmt.Errorf("failed to insert user: %v", err)
		}
		log.Printf("Created new user %s with internal ID: %s", user.NickName, user.UserID)
	} else if err != nil {
		return fmt.Errorf("failed to check user: %v", err)
	}

	return nil
}

func ProfileHandler(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, db *sql.DB) {
	// Retrieve session
	session, err := store.Get(r, "vivacity-session")
	if err != nil {
		log.Printf("Error retrieving session: %v", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	log.Printf("Session data: %v", session.Values)

	// Get userID
	var userID int64
	if val, ok := session.Values["UserID"]; ok && val != nil {
		switch v := val.(type) {
		case int64:
			userID = v
		case int:
			userID = int64(v)
		case string:
			if id, err := strconv.ParseInt(v, 10, 64); err == nil {
				userID = id
			} else {
				log.Printf("Invalid UserID in session on line 63: %v", v)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		default:
			log.Printf("Unsupported UserID type in session on line 68: %T", val)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	log.Printf("Fetching profile for UserID: %d", userID)

	// Fetch user data
	var u User
	err = db.QueryRow("SELECT username FROM users WHERE user_id = $1", userID).
		Scan(&u.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found for ID: %d", userID)
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("Database error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}
	log.Printf("User data: Username=%s", u.Username)

	var teamName string
	var battletag string
	var TeamID int
	if u.Team == "" {
		TeamID = 0
		log.Printf("User %s is not assigned to any team", u.Username)
		log.Printf("You are not assigned to any team. Please contact an administrator.")
	} else {
		TeamID, err = strconv.Atoi(u.Team)
		if err != nil {
			log.Printf("Invalid team ID: %v", err)
			http.Error(w, "Invalid team ID", http.StatusInternalServerError)
			return
		}
	}

	if TeamID == 0 {
		session.Values["TeamID"] = 0
		session.Values["TeamName"] = "No Team Assigned"
		battletag = session.Values["battletag"].(string)
		err = session.Save(r, w)
		if err != nil {
			log.Printf("Error saving session: %v", err)
			http.Error(w, "Session save error", http.StatusInternalServerError)
			return
		}
	} else {
		err = db.QueryRow("SELECT name FROM teams WHERE id = $1", TeamID).Scan(&teamName)
		if err != nil {
			log.Printf("Database error fetching team name: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		session.Values["TeamID"] = TeamID
		session.Values["TeamName"] = teamName
		battletag = session.Values["battletag"].(string)
		err = session.Save(r, w)
		if err != nil {
			log.Printf("Error saving session: %v", err)
			http.Error(w, "Session save error", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>Profile Page</title>
			<script src="https://cdn.tailwindcss.com"></script>
		</head>
		<body class="bg-gray-900 text-white flex items-center justify-center h-screen">
			<div class="text-center">
				<h1 class="text-4xl font-bold">Welcome, %s</h1>
				<p class="text-xl mt-2">Your Team is: %s</p>
				<a href="/" class="mt-4 inline-block bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded">Home</a>
			</div>
		</body>
		</html>
	`, battletag, session.Values["TeamName"])

	log.Printf("Successfully served profile page for user %s", userID)
}
