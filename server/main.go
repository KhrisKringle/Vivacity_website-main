package main

import (

	//"encoding/json"

	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/KhrisKringle/Vivacity_website-main/server/api"
	"github.com/KhrisKringle/Vivacity_website-main/server/datab"
	"github.com/KhrisKringle/Vivacity_website-main/server/user_account"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/battlenet"
)

var store *sessions.CookieStore

func init() {
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_KEY environment variable not set")
	}
	sessionKey, err := hex.DecodeString(sessionSecret)
	if err != nil {
		log.Fatal("Error decoding SESSION_KEY: ", err)
	}
	if len(sessionKey) != 32 {
		log.Fatal("SESSION_KEY must be 32 bytes (64 hex characters)")
	}

	// Initialize session store
	store = sessions.NewCookieStore(sessionKey)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600, // 1 hour for OAuth flow
		HttpOnly: true,
		Secure:   false,                // False for local dev
		SameSite: http.SameSiteLaxMode, // Allows OAuth redirects
	}

	bliz_secret := os.Getenv("BLIZZARD_CLIENT_SECRET")
	bliz_public := os.Getenv("BLIZZARD_PUBLIC")

	if bliz_public == "" || bliz_secret == "" {
		log.Fatal("BATTLENET_KEY, BATTLENET_SECRET must be set.")
	}

	// Add these lines for debugging
	log.Printf("SESSION_SECRET loaded: %t", os.Getenv("SESSION_SECRET") != "")
	log.Printf("BLIZZARD_PUBLIC loaded: %t", os.Getenv("BLIZZARD_PUBLIC") != "")
	log.Printf("BLIZZARD_CLIENT_SECRET loaded: %t", os.Getenv("BLIZZARD_CLIENT_SECRET") != "")

	// 3. Configure gothic to use your store and session name.
	gothic.Store = store

	callbackURL := "http://localhost:8080/auth/callback/battlenet"

	// Configure Gothic with Blizzard as the provider
	goth.UseProviders(
		battlenet.New(bliz_public, bliz_secret, callbackURL, "us"),
	)
	log.Println("Session store configured!")
}

func main() {

	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Connect to PostgreSQL
	db, err := datab.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Call the SetupDB function from the db package
	err = datab.SetupDB(db)
	if err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}

	fmt.Println("Database setup completed successfully!")

	r.Get("/auth/status", func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "_gothic_session")
		log.Printf("Auth status session values: %v", session.Values)
		if err != nil {
			log.Printf("Error getting session: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if authenticated, ok := session.Values["authenticated"].(bool); !ok || !authenticated {
			log.Printf("User is not authorized: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		user := map[string]string{
			"UserID":    session.Values["UserID"].(string),
			"battletag": session.Values["battletag"].(string),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		log.Printf("Auth status checked for user: %v", user)
	})

	// Add authentication routes
	r.Get("/auth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		// This is a small trick to tell Goth which provider to use.
		// It reads it from the context we set here.
		providerName := chi.URLParam(r, "provider")
		r = r.WithContext(context.WithValue(r.Context(), "provider", providerName))
		log.Println("Starting auth for provider:", providerName)

		// // Create a new session
		// session := sessions.NewSession(store, "_gothic_session")
		// session.Options = &sessions.Options{
		// 	Path:     "/",
		// 	MaxAge:   3600,
		// 	HttpOnly: true,
		// 	Secure:   false,
		// 	SameSite: http.SameSiteLaxMode,
		// }
		// session.Values["test"] = "test-value"
		// err := session.Save(r, w)
		// if err != nil {
		// 	log.Printf("Error saving session: %v", err)
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// log.Printf("Session initialized with test value: %v", session.Values["test"])
		for _, cookie := range r.Cookies() {
			log.Printf("Received cookie: %s = %s", cookie.Name, cookie.Value)
		}
		user, err := gothic.CompleteUserAuth(w, r)
		if err == nil {
			// User is already authenticated, no need to re-authenticate
			log.Printf("User already authenticated: %v", user)
			http.Redirect(w, r, fmt.Sprintf("/profile/%s", user.UserID), http.StatusFound)
			return
		} else {
			gothic.BeginAuthHandler(w, r)
		}
	})

	r.Get("/auth/callback/{provider}", func(w http.ResponseWriter, r *http.Request) {
		providerName := chi.URLParam(r, "provider")
		r = r.WithContext(context.WithValue(r.Context(), "provider", providerName))
		log.Println("Callback for provider:", providerName)

		// Log cookies
		for _, cookie := range r.Cookies() {
			log.Printf("Received cookie: %s = %s", cookie.Name, cookie.Value)
		}

		// session, err := store.Get(r, "_gothic_session")
		// if err != nil {
		// 	log.Printf("Session error: %v", err)
		// }
		// if testValue, ok := session.Values["test"]; ok {
		// 	log.Printf("Test value found in session: %v", testValue)
		// } else {
		// 	log.Println("Test value not found in session")
		// }
		// log.Printf("Session values: %v", session.Values)
		// log.Printf("Callback query params: %v", r.URL.Query())

		// Complete user authentication and obtain user data
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Printf("Error during callback for provider: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("USER AUTHENTICATED: %+v", user)

		// Process user data from Blizzard and create/update account and get back both the userID AND teamID
		err = user_account.HandleBlizzardAuth(db, user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to process user: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("User data processed for: %s", user.NickName)

		session, err := store.Get(r, "_gothic_session")
		if err != nil {
			log.Printf("Error getting session: %v", err)
			http.Error(w, fmt.Sprintf("Failed to get session: %v", err), http.StatusInternalServerError)
			return
		}

		session.Values["battletag"] = user.NickName
		session.Values["UserID"] = user.UserID
		session.Values["authenticated"] = true

		err = session.Save(r, w)
		if err != nil {
			log.Printf("Error saving session: %v", err)
			http.Error(w, fmt.Sprintf("Failed to save session: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Authentication successful for user: %v", user)
		// Redirect user to loggedin page
		http.Redirect(w, r, fmt.Sprintf("/profile/%s", session.Values["UserID"]), http.StatusFound)

	})

	// r.Get("/login", func(w http.ResponseWriter, r *http.Request) { // Test route to verify login

	// 	// Log cookies
	// 	for _, cookie := range r.Cookies() {
	// 		log.Printf("Received cookie: %s = %s", cookie.Name, cookie.Value)
	// 	}

	// 	// Verify session
	// 	session, err := store.Get(r, "_gothic_session")
	// 	if err != nil {
	// 		log.Printf("Error getting session: %v", err)
	// 	}
	// 	if authenticated, ok := session.Values["authenticated"].(bool); !ok || !authenticated {
	// 		log.Printf("User Not Authorized on line 241: %v", err)
	// 		log.Printf("Session values: %v", session.Values)
	// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 		return
	// 	}
	// 	log.Printf("Login route session values: %v", session.Values)
	// 	// Serve a simple logged-in page
	// 	http.Redirect(w, r, fmt.Sprintf("/profile/%s", session.Values["UserID"]), http.StatusFound)
	// })

	r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, gothic.SessionName)

		// Set authenticated to false and clear user data.
		session.Values["authenticated"] = false
		session.Values["battletag"] = ""
		session.Values["userID"] = ""
		err := gothic.Logout(w, r)
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
		if err != nil {
			log.Printf("Error during logout: %v", err)
			http.Error(w, "Error during logout", http.StatusInternalServerError)
			return
		}
		err = session.Save(r, w)
		if err != nil {
			log.Printf("Error saving session during logout: %v", err)
			http.Error(w, "Error saving session", http.StatusInternalServerError)
			return
		}
		log.Println("User logged out successfully")
		// Redirect to home page after logout
		http.Redirect(w, r, "/", http.StatusFound)
	})

	// Teams API
	r.Route("/api/teams", func(r chi.Router) {
		r.Post("/", api.TeamHandler(db))
		// Team-specific routes
		r.Route("/{team_id}", func(r chi.Router) {
			// Ensure teamID is an integer
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					teamID := chi.URLParam(r, "team_id")
					if _, err := strconv.Atoi(teamID); err != nil {
						http.Error(w, "Invalid team ID", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			})
			// Team-specific handlers
			r.Get("/", api.TeamHandler(db))    // Get team by ID
			r.Delete("/", api.TeamHandler(db)) // Delete team by ID
			r.Put("/", api.TeamHandler(db))    // Update team name by ID

			r.Get("/members", api.TeamMembersHandler(db))    // Get members of a team
			r.Post("/members", api.TeamMembersHandler(db))   // Add a member to a team
			r.Delete("/members", api.TeamMembersHandler(db)) // Remove a member from a team
			r.Put("/members", api.TeamMembersHandler(db))    // Update a member's role in a team

			r.Get("/schedule", api.ScheduleHandler(db))    // Get schedule for a team
			r.Post("/schedule", api.ScheduleHandler(db))   // Create schedule for a team
			r.Delete("/schedule", api.ScheduleHandler(db)) // Delete schedule for a team
			r.Put("/schedule", api.ScheduleHandler(db))    // Update schedule for a team

			r.Get("/availability", api.AvailabilityHandler(db))  // Get availability for a team
			r.Post("/availability", api.AvailabilityHandler(db)) // Set availability for a team
		})
	})

	// Players API
	r.Route("/api/players", func(r chi.Router) {
		r.Get("/", api.PlayerHandler(db)) // Get all players
		r.Route("/{user_id}", func(r chi.Router) {
			// Ensure userID is an integer
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					userID := chi.URLParam(r, "user_id")
					if _, err := strconv.Atoi(userID); err != nil {
						http.Error(w, "Invalid user ID", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			})
			// Player-specific handlers
			r.Get("/", api.PlayerHandler(db))
			r.Post("/", api.PlayerHandler(db))
			r.Delete("/", api.PlayerHandler(db))
			r.Put("/", api.PlayerHandler(db))
		})
	})

	// TimeSlots API
	r.Route("/api/timeslots", func(r chi.Router) {
		r.Get("/", api.TimeSlotsHandler(db)) // Get all time slot
		r.Route("/{timeSlotID}", func(r chi.Router) {
			// Ensure timeSlotID is an integer
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					timeSlotID := chi.URLParam(r, "timeSlotID")
					if _, err := strconv.Atoi(timeSlotID); err != nil {
						http.Error(w, "Invalid time slot ID", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			})
			// TimeSlot-specific handlers
			r.Get("/", api.TimeSlotsHandler(db))
			r.Post("/", api.TimeSlotsHandler(db))
			r.Delete("/", api.TimeSlotsHandler(db))
			r.Put("/", api.TimeSlotsHandler(db))
		})
	})

	// Serve static files (CSS, JS, images)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("../static"))))

	// Routes for HTML pages
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../static/index.html")
	})

	r.Get("/teams", func(w http.ResponseWriter, r *http.Request) {
		sessions, err := store.Get(r, "_gothic_session")
		if err != nil {
			log.Printf("Error getting session: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		log.Printf("Teams route session values: %v", sessions.Values)
		http.ServeFile(w, r, "../static/Teams/teams.html")
	})

	// This route serves the team profile page
	r.Get("/team-profile", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../static/TeamProfilePage/team_page.html")
	})

	r.Get("/schedule", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../static/Scheduling_Page/scheduling_page.html")
	})

	// Routes for Profile
	r.Route("/profile", func(r chi.Router) {
		r.Get("/{user_id}", func(w http.ResponseWriter, r *http.Request) {
			user_account.ProfileHandler(w, r, store, db)
		})
	})

	// Start server
	log.Println("Server starting on :8080")
	http.ListenAndServe(":8080", r)
}
