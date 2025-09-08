package user_account

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
)

type User struct {
	Username string
	Team     string
}

// and returns their internal database ID and their team ID.
func HandleBlizzardAuth(db *sql.DB, user goth.User) (userID int64, teamID sql.NullInt64, err error) {
	// Check if the user already exists in the database by their Blizzard UserID
	blizzardUserID, _ := strconv.ParseInt(user.UserID, 10, 64)
	query := "SELECT id FROM users WHERE user_id = $1" // Assuming 'id' is your primary key
	err = db.QueryRow(query, blizzardUserID).Scan(&userID)

	if err == sql.ErrNoRows {
		// If the user does not exist, insert a new record and get their new primary key ID
		insertQuery := "INSERT INTO users (username, user_id) VALUES ($1, $2) RETURNING id"
		err = db.QueryRow(insertQuery, user.NickName, blizzardUserID).Scan(&userID)
		if err != nil {
			return 0, sql.NullInt64{}, fmt.Errorf("failed to insert user: %v", err)
		}
	} else if err != nil {
		return 0, sql.NullInt64{}, fmt.Errorf("failed to check user: %v", err)
	}

	// Now that we have the user's internal ID, find their team_id.
	// This query finds the team_id for the user. We use sql.NullInt64
	// to correctly handle cases where a user is not on a team.
	teamQuery := "SELECT team_id FROM team_members WHERE user_id = $1 LIMIT 1"
	err = db.QueryRow(teamQuery, userID).Scan(&teamID)
	if err != nil && err != sql.ErrNoRows {
		// An actual error occurred (not just 'no team found')
		return userID, sql.NullInt64{}, fmt.Errorf("failed to query for team_id: %v", err)
	}

	// If the user is not on a team, err will be sql.ErrNoRows, and teamID.Valid will be false.
	// This is the correct behavior.
	return userID, teamID, nil
}

func GenerateSessionSecret(length int) (string, error) {
	// Create a byte slice to hold the random bytes
	secret := make([]byte, length)

	// Fill the slice with cryptographically secure random bytes
	_, err := rand.Read(secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// Encode the bytes to a base64 string for easy usage
	return base64.StdEncoding.EncodeToString(secret), nil
}

func ProfileHandler(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, db *sql.DB) {
	// Retrieve session
	session, err := store.Get(r, "auth-session")
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
				log.Printf("Invalid UserID in session: %v", v)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
		default:
			log.Printf("Unsupported UserID type in session: %T", val)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
	} else {
		// Fallback to URL param
		urlUserID := chi.URLParam(r, "UserID")
		if urlUserID == "" {
			log.Println("No UserID in session or URL")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if id, err := strconv.ParseInt(urlUserID, 10, 64); err == nil {
			userID = id
		} else {
			log.Printf("Invalid UserID in URL: %s", urlUserID)
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
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

	// Render template
	tmpl, err := template.ParseFiles("../static/Profile/profile.html")
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, u); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
	}
}
