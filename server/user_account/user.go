package user_account

import (
	"database/sql"
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
	} else if err != nil {
		return fmt.Errorf("failed to check user: %v", err)
	}

	return nil
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
