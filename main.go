package main

import (
	"log"
	"net/http"

	"evoting-app/internal/config"
	"evoting-app/internal/database"
	"evoting-app/internal/handlers"
	"evoting-app/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize session store
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))

	// Set session store for middleware
	middleware.SetSessionStore(store)

	// Initialize handlers
	h := handlers.New(db, store)

	// Setup routes
	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	// Public routes
	r.HandleFunc("/", h.Home).Methods("GET")
	r.HandleFunc("/login", h.Login).Methods("GET", "POST")
	r.HandleFunc("/vote", h.VoteForm).Methods("GET")
	r.HandleFunc("/vote", h.SubmitVote).Methods("POST")
	r.HandleFunc("/logout", h.Logout).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/admin").Subrouter()
	protected.Use(middleware.RequireAuth)

	// Superadmin routes
	superadmin := protected.PathPrefix("/superadmin").Subrouter()
	superadmin.Use(middleware.RequireSuperAdmin)
	superadmin.HandleFunc("/dashboard", h.SuperAdminDashboard).Methods("GET")
	superadmin.HandleFunc("/elections", h.ManageElections).Methods("GET")
	superadmin.HandleFunc("/elections/create", h.CreateElection).Methods("GET", "POST")
	superadmin.HandleFunc("/elections/{id}/edit", h.EditElection).Methods("GET", "POST")
	superadmin.HandleFunc("/elections/{id}/delete", h.DeleteElection).Methods("POST")
	superadmin.HandleFunc("/elections/{id}/assign-admin", h.AssignAdmin).Methods("GET", "POST")
	superadmin.HandleFunc("/users", h.ManageUsers).Methods("GET")
	superadmin.HandleFunc("/users/create", h.CreateUser).Methods("GET", "POST")

	// Admin routes
	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.RequireAdmin)
	admin.HandleFunc("/dashboard", h.AdminDashboard).Methods("GET")
	admin.HandleFunc("/elections", h.AdminElections).Methods("GET")
	admin.HandleFunc("/elections/{id}/candidates", h.ManageCandidates).Methods("GET")
	admin.HandleFunc("/elections/{id}/candidates/create", h.CreateCandidate).Methods("GET", "POST")
	admin.HandleFunc("/elections/{id}/candidates/{candidate_id}/edit", h.EditCandidate).Methods("GET", "POST")
	admin.HandleFunc("/elections/{id}/candidates/{candidate_id}/delete", h.DeleteCandidate).Methods("POST")
	admin.HandleFunc("/elections/{id}/tokens", h.ManageTokens).Methods("GET")
	admin.HandleFunc("/elections/{id}/tokens/generate", h.GenerateTokens).Methods("POST")
	admin.HandleFunc("/elections/{id}/votes", h.ManageVotes).Methods("GET")
	admin.HandleFunc("/elections/{id}/reports", h.ElectionReports).Methods("GET")

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
