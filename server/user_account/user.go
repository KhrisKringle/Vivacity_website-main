package user_account

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	} else {
		log.Printf("User already exists with ID: %s", user.UserID)
	}

	return nil
}

func ProfileHandler(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, db *sql.DB) {
	// Retrieve session
	session, err := store.Get(r, "_gothic_session")
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
				log.Printf("Invalid UserID in session on line 65: %v", v)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		default:
			log.Printf("Unsupported UserID type in session on line 70: %T", val)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	} else {
		// Fallback to URL param
		urlUserID := chi.URLParam(r, "UserID")
		if urlUserID == "" {
			log.Println("No UserID in session or URL on line 78")
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		if id, err := strconv.ParseInt(urlUserID, 10, 64); err == nil {
			userID = id
		} else {
			log.Printf("Invalid UserID in URLline 85: %s", urlUserID)
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

	battletag := session.Values["battletag"].(string)

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
				<p class="text-xl mt-2">Your User ID is: %s</p>
				<a href="/" class="mt-4 inline-block bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded">Home</a>
			</div>
		</body>
		</html>
	`, battletag, userID)

	log.Printf("Successfully served profile page for user %s", userID)
}
