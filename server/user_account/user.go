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

// HandleBlizzardAuth processes the authenticated Blizzard user data and creates or updates their account in the database.
func HandleBlizzardAuth(db *sql.DB, user goth.User) error {
	// Check if the user already exists in the database
	var userID int64
	query := "SELECT user_id FROM users WHERE username = $1"
	err := db.QueryRow(query, user.NickName).Scan(&userID)
	if err == sql.ErrNoRows {
		// If the user does not exist, insert a new record
		userIDInt, err := strconv.ParseInt(user.UserID, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse user ID: %v", err)
		}
		_, err = db.Exec("INSERT INTO users (username, user_id) VALUES ($1, $2)", user.NickName, userIDInt)
		if err != nil {
			return fmt.Errorf("failed to insert user: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check user: %v", err)
	}

	// You can add additional logic here to update the user if needed
	return nil
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
