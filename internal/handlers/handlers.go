package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"evoting-app/internal/middleware"
	"evoting-app/internal/models"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	db    *sql.DB
	store sessions.Store
	tmpl  *template.Template
	auth  *middleware.AuthService
}

func New(db *sql.DB, store sessions.Store) *Handlers {
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

	// Load templates with custom functions
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob(filepath.Join("web", "templates", "*.html")))

	// Debug: list all templates
	log.Printf("Loaded templates: %v", tmpl.DefinedTemplates())

	return &Handlers{
		db:    db,
		store: store,
		tmpl:  tmpl,
		auth:  middleware.NewAuthService(db),
	}
}

// Helper function to render templates correctly
func (h *Handlers) renderTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
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
		filepath.Join("web", "templates", "base.html"),
		filepath.Join("web", "templates", templateName),
	)
	if err != nil {
		log.Printf("Error parsing templates: %v", err)
		return err
	}

	return tmpl.ExecuteTemplate(w, "base.html", data)
}

// Home page
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	log.Printf("Home handler called for path: %s", r.URL.Path)

	err := h.renderTemplate(w, "home.html", nil)
	if err != nil {
		log.Printf("Error executing home template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Login handlers
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Check if user is already logged in
		session, _ := h.store.Get(r, "session")
		if userID, ok := session.Values["user_id"]; ok {
			// User is already logged in, redirect to appropriate dashboard
			if role, roleOk := session.Values["role"]; roleOk {
				if role == "superadmin" {
					http.Redirect(w, r, "/admin/superadmin/dashboard", http.StatusSeeOther)
					return
				} else if role == "admin" {
					http.Redirect(w, r, "/admin/admin/dashboard", http.StatusSeeOther)
					return
				}
			}
			log.Printf("User %v already logged in, redirecting", userID)
		}

		err := h.renderTemplate(w, "login.html", nil)
		if err != nil {
			log.Printf("Error executing login template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.auth.GetUserByUsername(username)
	if err != nil {
		h.renderTemplate(w, "login.html", map[string]string{
			"Error": "Invalid username or password",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		h.renderTemplate(w, "login.html", map[string]string{
			"Error": "Invalid username or password",
		})
		return
	}

	session, _ := h.store.Get(r, "session")
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Values["role"] = user.Role
	session.Save(r, w)

	if user.Role == "superadmin" {
		http.Redirect(w, r, "/admin/superadmin/dashboard", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/admin/dashboard", http.StatusSeeOther)
	}
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "session")
	session.Values = make(map[interface{}]interface{})
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Voting handlers
func (h *Handlers) VoteForm(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		err := h.renderTemplate(w, "vote_token.html", nil)
		if err != nil {
			log.Printf("Error executing vote token template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Validate token and get election
	election, candidates, err := h.getElectionByToken(token)
	if err != nil {
		err = h.renderTemplate(w, "vote_token.html", map[string]string{
			"Error": "Invalid or expired token",
		})
		if err != nil {
			log.Printf("Error executing vote token template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	data := map[string]interface{}{
		"Election":   election,
		"Candidates": candidates,
		"Token":      token,
	}

	err = h.renderTemplate(w, "vote_form.html", data)
	if err != nil {
		log.Printf("Error executing vote form template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) SubmitVote(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	candidateID := r.FormValue("candidate_id")

	// Validate token
	tokenRecord, err := h.getTokenRecord(token)
	if err != nil || tokenRecord.IsUsed {
		err = h.renderTemplate(w, "vote_result.html", map[string]interface{}{
			"Success": false,
			"Message": "Invalid or already used token",
		})
		if err != nil {
			log.Printf("Error executing vote result template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Submit vote
	err = h.submitVote(tokenRecord.ID, tokenRecord.ElectionID, candidateID)
	if err != nil {
		err = h.renderTemplate(w, "vote_result.html", map[string]interface{}{
			"Success": false,
			"Message": "Failed to submit vote",
		})
		if err != nil {
			log.Printf("Error executing vote result template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	err = h.renderTemplate(w, "vote_result.html", map[string]interface{}{
		"Success": true,
		"Message": "Vote submitted successfully",
	})
	if err != nil {
		log.Printf("Error executing vote result template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Helper functions
func (h *Handlers) getElectionByToken(token string) (*models.Election, []models.Candidate, error) {
	// Get election from token
	query := `
		SELECT e.id, e.title, e.description, e.start_date, e.end_date, e.status
		FROM elections e
		JOIN voting_tokens vt ON e.id = vt.election_id
		WHERE vt.token = ? AND vt.is_used = FALSE AND e.status = 'active'
	`

	election := &models.Election{}
	err := h.db.QueryRow(query, token).Scan(
		&election.ID, &election.Title, &election.Description,
		&election.StartDate, &election.EndDate, &election.Status,
	)
	if err != nil {
		return nil, nil, err
	}

	// Get candidates
	candidatesQuery := `SELECT id, name, description, photo_url FROM candidates WHERE election_id = ? ORDER BY order_num`
	rows, err := h.db.Query(candidatesQuery, election.ID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var candidates []models.Candidate
	for rows.Next() {
		var candidate models.Candidate
		err := rows.Scan(&candidate.ID, &candidate.Name, &candidate.Description, &candidate.PhotoURL)
		if err != nil {
			return nil, nil, err
		}
		candidates = append(candidates, candidate)
	}

	return election, candidates, nil
}

func (h *Handlers) getTokenRecord(token string) (*models.VotingToken, error) {
	tokenRecord := &models.VotingToken{}
	query := `SELECT id, election_id, token, is_used FROM voting_tokens WHERE token = ?`

	err := h.db.QueryRow(query, token).Scan(
		&tokenRecord.ID, &tokenRecord.ElectionID, &tokenRecord.Token, &tokenRecord.IsUsed,
	)

	return tokenRecord, err
}

func (h *Handlers) submitVote(tokenID int, electionID int, candidateID string) error {
	tx, err := h.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert vote
	_, err = tx.Exec(
		`INSERT INTO votes (election_id, candidate_id, token_id) VALUES (?, ?, ?)`,
		electionID, candidateID, tokenID,
	)
	if err != nil {
		return err
	}

	// Mark token as used
	_, err = tx.Exec(
		`UPDATE voting_tokens SET is_used = TRUE, used_at = CURRENT_TIMESTAMP WHERE id = ?`,
		tokenID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
