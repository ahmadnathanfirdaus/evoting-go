package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"evoting-app/internal/middleware"
	"evoting-app/internal/models"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// Helper function to render templates correctly
func (h *Handlers) renderSuperAdminTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
	// Create function map for templates
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"max": func(a, b int) int {
			if a > b {
				return a
			}
			return b
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFiles(
		filepath.Join("web", "templates", "admin_base.html"),
		filepath.Join("web", "templates", templateName),
	)
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		return err
	}

	return tmpl.ExecuteTemplate(w, "admin_base.html", data)
}

// SuperAdmin Dashboard
func (h *Handlers) SuperAdminDashboard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	// Get statistics
	stats := h.getSuperAdminStats()

	data := map[string]interface{}{
		"User":  user,
		"Stats": stats,
	}

	err := h.renderSuperAdminTemplate(w, "superadmin_dashboard.html", data)
	if err != nil {
		log.Printf("Error executing superadmin dashboard template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Election Management
func (h *Handlers) ManageElections(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	elections, err := h.getAllElections()
	if err != nil {
		http.Error(w, "Failed to load elections", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":      user,
		"Elections": elections,
	}

	err = h.renderSuperAdminTemplate(w, "manage_elections.html", data)
	if err != nil {
		log.Printf("Error executing manage elections template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) CreateElection(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	if r.Method == "GET" {
		err := h.renderSuperAdminTemplate(w, "create_election.html", map[string]interface{}{
			"User": user,
		})
		if err != nil {
			log.Printf("Error executing create election template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Handle POST
	title := r.FormValue("title")
	description := r.FormValue("description")
	startDate := r.FormValue("start_date")
	endDate := r.FormValue("end_date")

	// Parse dates
	start, err := time.Parse("2006-01-02T15:04", startDate)
	if err != nil {
		h.renderSuperAdminTemplate(w, "create_election.html", map[string]interface{}{
			"User":  user,
			"Error": "Invalid start date format",
		})
		return
	}

	end, err := time.Parse("2006-01-02T15:04", endDate)
	if err != nil {
		h.renderSuperAdminTemplate(w, "create_election.html", map[string]interface{}{
			"User":  user,
			"Error": "Invalid end date format",
		})
		return
	}

	// Create election
	_, err = h.db.Exec(
		`INSERT INTO elections (title, description, start_date, end_date, created_by) VALUES (?, ?, ?, ?, ?)`,
		title, description, start, end, user.ID,
	)

	if err != nil {
		h.renderSuperAdminTemplate(w, "create_election.html", map[string]interface{}{
			"User":  user,
			"Error": "Failed to create election",
		})
		return
	}

	http.Redirect(w, r, "/admin/superadmin/elections", http.StatusSeeOther)
}

func (h *Handlers) EditElection(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	vars := mux.Vars(r)
	electionID := vars["id"]

	if r.Method == "GET" {
		election, err := h.getElectionByID(electionID)
		if err != nil {
			http.Error(w, "Election not found", http.StatusNotFound)
			return
		}

		err = h.renderSuperAdminTemplate(w, "edit_election.html", map[string]interface{}{
			"User":     user,
			"Election": election,
		})
		if err != nil {
			log.Printf("Error executing edit election template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Handle POST
	title := r.FormValue("title")
	description := r.FormValue("description")
	startDate := r.FormValue("start_date")
	endDate := r.FormValue("end_date")
	status := r.FormValue("status")

	start, err := time.Parse("2006-01-02T15:04", startDate)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}

	end, err := time.Parse("2006-01-02T15:04", endDate)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec(
		`UPDATE elections SET title = ?, description = ?, start_date = ?, end_date = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		title, description, start, end, status, electionID,
	)

	if err != nil {
		http.Error(w, "Failed to update election", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/superadmin/elections", http.StatusSeeOther)
}

func (h *Handlers) DeleteElection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	electionID := vars["id"]

	_, err := h.db.Exec(`DELETE FROM elections WHERE id = ?`, electionID)
	if err != nil {
		http.Error(w, "Failed to delete election", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/superadmin/elections", http.StatusSeeOther)
}

// Admin Assignment
func (h *Handlers) AssignAdmin(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	vars := mux.Vars(r)
	electionID := vars["id"]

	if r.Method == "GET" {
		election, err := h.getElectionByID(electionID)
		if err != nil {
			http.Error(w, "Election not found", http.StatusNotFound)
			return
		}

		admins, err := h.getAllAdmins()
		if err != nil {
			http.Error(w, "Failed to load admins", http.StatusInternalServerError)
			return
		}

		assignedAdmins, err := h.getAssignedAdmins(electionID)
		if err != nil {
			http.Error(w, "Failed to load assigned admins", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"User":           user,
			"Election":       election,
			"Admins":         admins,
			"AssignedAdmins": assignedAdmins,
		}

		err = h.renderSuperAdminTemplate(w, "assign_admin.html", data)
		if err != nil {
			log.Printf("Error executing assign admin template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Handle POST
	adminID := r.FormValue("admin_id")

	_, err := h.db.Exec(
		`INSERT OR IGNORE INTO election_admins (election_id, user_id) VALUES (?, ?)`,
		electionID, adminID,
	)

	if err != nil {
		http.Error(w, "Failed to assign admin", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/superadmin/elections/"+electionID+"/assign-admin", http.StatusSeeOther)
}

// User Management
func (h *Handlers) ManageUsers(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	users, err := h.getAllUsers()
	if err != nil {
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":  user,
		"Users": users,
	}

	err = h.renderSuperAdminTemplate(w, "manage_users.html", data)
	if err != nil {
		log.Printf("Error executing manage users template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	if r.Method == "GET" {
		err := h.renderSuperAdminTemplate(w, "create_user.html", map[string]interface{}{
			"User": user,
		})
		if err != nil {
			log.Printf("Error executing create user template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Handle POST
	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		h.renderSuperAdminTemplate(w, "create_user.html", map[string]interface{}{
			"User":  user,
			"Error": "Failed to hash password",
		})
		return
	}

	_, err = h.db.Exec(
		`INSERT INTO users (username, password, role) VALUES (?, ?, ?)`,
		username, string(hashedPassword), role,
	)

	if err != nil {
		h.renderSuperAdminTemplate(w, "create_user.html", map[string]interface{}{
			"User":  user,
			"Error": "Failed to create user",
		})
		return
	}

	http.Redirect(w, r, "/admin/superadmin/users", http.StatusSeeOther)
}

// Helper functions
func (h *Handlers) getSuperAdminStats() map[string]int {
	stats := make(map[string]int)

	var totalElections, totalAdmins, activeElections, totalVotes int
	h.db.QueryRow("SELECT COUNT(*) FROM elections").Scan(&totalElections)
	h.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&totalAdmins)
	h.db.QueryRow("SELECT COUNT(*) FROM elections WHERE status = 'active'").Scan(&activeElections)
	h.db.QueryRow("SELECT COUNT(*) FROM votes").Scan(&totalVotes)

	stats["total_elections"] = totalElections
	stats["total_admins"] = totalAdmins
	stats["active_elections"] = activeElections
	stats["total_votes"] = totalVotes

	return stats
}

func (h *Handlers) getAllElections() ([]models.Election, error) {
	query := `SELECT id, title, description, start_date, end_date, status, created_at FROM elections ORDER BY created_at DESC`
	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var elections []models.Election
	for rows.Next() {
		var election models.Election
		err := rows.Scan(
			&election.ID, &election.Title, &election.Description,
			&election.StartDate, &election.EndDate, &election.Status, &election.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		elections = append(elections, election)
	}

	return elections, nil
}

func (h *Handlers) getElectionByID(id string) (*models.Election, error) {
	election := &models.Election{}
	query := `SELECT id, title, description, start_date, end_date, status, created_at FROM elections WHERE id = ?`

	err := h.db.QueryRow(query, id).Scan(
		&election.ID, &election.Title, &election.Description,
		&election.StartDate, &election.EndDate, &election.Status, &election.CreatedAt,
	)

	return election, err
}

func (h *Handlers) getAllAdmins() ([]models.User, error) {
	query := `SELECT id, username, role, created_at FROM users WHERE role = 'admin' ORDER BY username`
	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admins []models.User
	for rows.Next() {
		var admin models.User
		err := rows.Scan(&admin.ID, &admin.Username, &admin.Role, &admin.CreatedAt)
		if err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}

	return admins, nil
}

func (h *Handlers) getAllUsers() ([]models.User, error) {
	query := `SELECT id, username, role, created_at FROM users ORDER BY created_at DESC`
	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (h *Handlers) getAssignedAdmins(electionID string) ([]models.User, error) {
	query := `
		SELECT u.id, u.username, u.role, ea.assigned_at 
		FROM users u 
		JOIN election_admins ea ON u.id = ea.user_id 
		WHERE ea.election_id = ? 
		ORDER BY ea.assigned_at DESC
	`
	rows, err := h.db.Query(query, electionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var admins []models.User
	for rows.Next() {
		var admin models.User
		var assignedAt time.Time
		err := rows.Scan(&admin.ID, &admin.Username, &admin.Role, &assignedAt)
		if err != nil {
			return nil, err
		}
		admins = append(admins, admin)
	}

	return admins, nil
}
