package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"Vivacity_website/server/api"
	"Vivacity_website/server/datab"
	"Vivacity_website/server/user_account"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"

	"github.com/markbates/goth/gothic"
)

var (
	sessionSecret, _ = user_account.GenerateSessionSecret(32)
	store            = sessions.NewCookieStore([]byte(sessionSecret))
)

// func getBlizzardSecret() string {
// 	secret := os.Getenv("BLIZZARD_CLIENT_SECRET")
// 	if secret == "" {
// 		log.Fatal("BLIZZARD_CLIENT_SECRET environment variable not set")
// 	}
// 	return secret
// }

func main() {
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

	// Generate a session secret of 32 bytes

	if err != nil {
		log.Fatalf("Error generating session secret: %v", err)
	}
	fmt.Println(sessionSecret)

	// Use a custom session store
	gothic.Store = store

	// Continue with your application setup
	log.Println("Session store configured!")

	// Configure Gothic with Blizzard as the provider
	// goth.UseProviders(
	// 	battlenet.New("a54315d3ec30453f9e58d1173caa05f6", getBlizzardSecret(), "http://192.168.1.234:8080/auth/callback/battlenet", "us"), // Adjust region as needed
	// )

	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// r.Use(middleware.Throttle(10)) // Allow 10 requests per second

	r.Get("/auth/user", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth-session")
		user, ok := session.Values["User"]
		if !ok {
			http.Error(w, "No user", http.StatusUnauthorized)
			return
		}
		if err := json.NewEncoder(w).Encode(&user); err != nil {
			http.Error(w, "I'm hurting on Line 75", http.StatusInternalServerError)
			return
		}
	})

	// Add authentication routes
	r.Get("/auth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		provider := chi.URLParam(r, "provider")
		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
		log.Printf("Starting authentication for provider: %s", provider)
		gothic.BeginAuthHandler(w, r)
	})

	r.Get("/auth/callback/{provider}", func(w http.ResponseWriter, r *http.Request) {
		provider := chi.URLParam(r, "provider")
		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Printf("Error during callback for provider: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Process user data from Blizzard and create/update account
		err = user_account.HandleBlizzardAuth(db, user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to process user: %v", err), http.StatusInternalServerError)
			return
		}

		// Store user in session
		session, _ := store.Get(r, "auth-session")
		session.Values["User"] = user          // Convert int to string
		session.Values["UserID"] = user.UserID // Store battletag separately
		session.Save(r, w)

		// Redirect user to profile page
		http.Redirect(w, r, fmt.Sprintf("/profile/%s", user.UserID), http.StatusFound)
		log.Printf("Authentication successful for user: %v", user)
	})

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../static/Profile/login.html")
	})

	r.Get("/logout/{provider}", func(w http.ResponseWriter, r *http.Request) {
		provider := chi.URLParam(r, "provider")
		r = r.WithContext(context.WithValue(r.Context(), "provider", provider))
		gothic.Logout(w, r)
		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	// Availability API
	r.Route("/api/availability", func(r chi.Router) {
		r.Get("/", api.AvailabilityHandler(db))
		r.Post("/", api.AvailabilityHandler(db))
	})

	// Teams API
	r.Route("/api/teams/", func(r chi.Router) {
		r.Post("/", api.TeamHandler(db))
		// Team-specific routes
		r.Route("/{teamID}", func(r chi.Router) {
			// Ensure teamID is an integer
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					teamID := chi.URLParam(r, "teamID")
					if _, err := strconv.Atoi(teamID); err != nil {
						http.Error(w, "Invalid team ID", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			})
			// Team-specific handlers
			r.Get("/", api.TeamHandler(db)) // Get team by ID
			r.Delete("/", api.TeamHandler(db))
			r.Put("/", api.TeamHandler(db))                  // Update team name by ID
			r.Get("/members", api.TeamMembersHandler(db))    // Get members of a team
			r.Post("/members", api.TeamMembersHandler(db))   // Add a member to a team
			r.Delete("/members", api.TeamMembersHandler(db)) // Remove a member from a team
			r.Put("/members", api.TeamMembersHandler(db))    // Update a member's role in a team
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
		http.ServeFile(w, r, "../static/Teams/teams.html")
	})

	// Routes for Profile
	r.Route("/profile", func(r chi.Router) {
		r.Get("/{UserID}", func(w http.ResponseWriter, r *http.Request) {
			user_account.ProfileHandler(w, r, store, db)
		})
	})

	// Start server
	http.ListenAndServe(":8080", r)
}
