package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"evoting-app/internal/middleware"
	"evoting-app/internal/models"

	"github.com/gorilla/mux"
)

// Helper function to render templates correctly for admin
func (h *Handlers) renderAdminTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
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

// Admin Dashboard
func (h *Handlers) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	// Get elections assigned to this admin
	elections, err := h.getAdminElections(user.ID)
	if err != nil {
		http.Error(w, "Failed to load elections", http.StatusInternalServerError)
		return
	}

	// Get statistics
	stats := h.getAdminStats(user.ID)

	data := map[string]interface{}{
		"User":      user,
		"Elections": elections,
		"Stats":     stats,
	}

	err = h.renderAdminTemplate(w, "admin_dashboard.html", data)
	if err != nil {
		log.Printf("Error executing admin dashboard template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Admin Elections
func (h *Handlers) AdminElections(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())

	elections, err := h.getAdminElections(user.ID)
	if err != nil {
		http.Error(w, "Failed to load elections", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":      user,
		"Elections": elections,
	}

	err = h.renderAdminTemplate(w, "admin_elections.html", data)
	if err != nil {
		log.Printf("Error executing admin elections template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Candidate Management
func (h *Handlers) ManageCandidates(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	vars := mux.Vars(r)
	electionID := vars["id"]

	// Check if admin has access to this election
	if !h.hasElectionAccess(user.ID, electionID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	election, err := h.getElectionByID(electionID)
	if err != nil {
		http.Error(w, "Election not found", http.StatusNotFound)
		return
	}

	candidates, err := h.getCandidatesByElection(electionID)
	if err != nil {
		http.Error(w, "Failed to load candidates", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":       user,
		"Election":   election,
		"Candidates": candidates,
	}

	err = h.renderAdminTemplate(w, "manage_candidates.html", data)
	if err != nil {
		log.Printf("Error executing manage candidates template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) CreateCandidate(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	vars := mux.Vars(r)
	electionID := vars["id"]

	if !h.hasElectionAccess(user.ID, electionID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		election, err := h.getElectionByID(electionID)
		if err != nil {
			http.Error(w, "Election not found", http.StatusNotFound)
			return
		}

		err = h.renderAdminTemplate(w, "create_candidate.html", map[string]interface{}{
			"User":     user,
			"Election": election,
		})
		if err != nil {
			log.Printf("Error executing create candidate template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Handle POST
	name := r.FormValue("name")
	description := r.FormValue("description")
	photoURL := r.FormValue("photo_url")
	orderStr := r.FormValue("order")

	order := 0
	if orderStr != "" {
		order, _ = strconv.Atoi(orderStr)
	}

	_, err := h.db.Exec(
		`INSERT INTO candidates (election_id, name, description, photo_url, order_num) VALUES (?, ?, ?, ?, ?)`,
		electionID, name, description, photoURL, order,
	)

	if err != nil {
		http.Error(w, "Failed to create candidate", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/admin/elections/"+electionID+"/candidates", http.StatusSeeOther)
}

func (h *Handlers) EditCandidate(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	vars := mux.Vars(r)
	electionID := vars["id"]
	candidateID := vars["candidate_id"]

	if !h.hasElectionAccess(user.ID, electionID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		candidate, err := h.getCandidateByID(candidateID)
		if err != nil {
			http.Error(w, "Candidate not found", http.StatusNotFound)
			return
		}

		election, err := h.getElectionByID(electionID)
		if err != nil {
			http.Error(w, "Election not found", http.StatusNotFound)
			return
		}

		err = h.renderAdminTemplate(w, "edit_candidate.html", map[string]interface{}{
			"User":      user,
			"Election":  election,
			"Candidate": candidate,
		})
		if err != nil {
			log.Printf("Error executing edit candidate template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Handle POST
	name := r.FormValue("name")
	description := r.FormValue("description")
	photoURL := r.FormValue("photo_url")
	orderStr := r.FormValue("order")

	order := 0
	if orderStr != "" {
		order, _ = strconv.Atoi(orderStr)
	}

	_, err := h.db.Exec(
		`UPDATE candidates SET name = ?, description = ?, photo_url = ?, order_num = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, description, photoURL, order, candidateID,
	)

	if err != nil {
		http.Error(w, "Failed to update candidate", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/admin/elections/"+electionID+"/candidates", http.StatusSeeOther)
}

func (h *Handlers) DeleteCandidate(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	vars := mux.Vars(r)
	electionID := vars["id"]
	candidateID := vars["candidate_id"]

	if !h.hasElectionAccess(user.ID, electionID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err := h.db.Exec(`DELETE FROM candidates WHERE id = ?`, candidateID)
	if err != nil {
		http.Error(w, "Failed to delete candidate", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/admin/elections/"+electionID+"/candidates", http.StatusSeeOther)
}

// Token Management
func (h *Handlers) ManageTokens(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	vars := mux.Vars(r)
	electionID := vars["id"]

	if !h.hasElectionAccess(user.ID, electionID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	election, err := h.getElectionByID(electionID)
	if err != nil {
		http.Error(w, "Election not found", http.StatusNotFound)
		return
	}

	tokens, err := h.getTokensByElection(electionID)
	if err != nil {
		http.Error(w, "Failed to load tokens", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":     user,
		"Election": election,
		"Tokens":   tokens,
	}

	err = h.renderAdminTemplate(w, "manage_tokens.html", data)
	if err != nil {
		log.Printf("Error executing manage tokens template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) GenerateTokens(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	vars := mux.Vars(r)
	electionID := vars["id"]

	if !h.hasElectionAccess(user.ID, electionID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	countStr := r.FormValue("count")
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 || count > 1000 {
		http.Error(w, "Invalid token count", http.StatusBadRequest)
		return
	}

	// Generate tokens
	for i := 0; i < count; i++ {
		token := generateRandomToken()
		_, err := h.db.Exec(
			`INSERT INTO voting_tokens (election_id, token) VALUES (?, ?)`,
			electionID, token,
		)
		if err != nil {
			http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/admin/admin/elections/"+electionID+"/tokens", http.StatusSeeOther)
}

// Vote Management
func (h *Handlers) ManageVotes(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	vars := mux.Vars(r)
	electionID := vars["id"]

	if !h.hasElectionAccess(user.ID, electionID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	election, err := h.getElectionByID(electionID)
	if err != nil {
		http.Error(w, "Election not found", http.StatusNotFound)
		return
	}

	votes, err := h.getVotesByElection(electionID)
	if err != nil {
		http.Error(w, "Failed to load votes", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":     user,
		"Election": election,
		"Votes":    votes,
	}

	err = h.renderAdminTemplate(w, "manage_votes.html", data)
	if err != nil {
		log.Printf("Error executing manage votes template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Reports
func (h *Handlers) ElectionReports(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	vars := mux.Vars(r)
	electionID := vars["id"]

	if !h.hasElectionAccess(user.ID, electionID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	election, err := h.getElectionByID(electionID)
	if err != nil {
		http.Error(w, "Election not found", http.StatusNotFound)
		return
	}

	voteCounts, err := h.getVoteCountsByElection(electionID)
	if err != nil {
		http.Error(w, "Failed to load vote counts", http.StatusInternalServerError)
		return
	}

	stats, err := h.getElectionStats(electionID)
	if err != nil {
		http.Error(w, "Failed to load election stats", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":       user,
		"Election":   election,
		"VoteCounts": voteCounts,
		"Stats":      stats,
	}

	err = h.renderAdminTemplate(w, "election_reports.html", data)
	if err != nil {
		log.Printf("Error executing election reports template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Helper functions
func (h *Handlers) getAdminElections(userID int) ([]models.Election, error) {
	query := `
		SELECT DISTINCT e.id, e.title, e.description, e.start_date, e.end_date, e.status, e.created_at
		FROM elections e
		JOIN election_admins ea ON e.id = ea.election_id
		WHERE ea.user_id = ?
		ORDER BY e.created_at DESC
	`
	rows, err := h.db.Query(query, userID)
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

func (h *Handlers) hasElectionAccess(userID int, electionID string) bool {
	// Check if user is superadmin first
	var role string
	userQuery := `SELECT role FROM users WHERE id = ?`
	err := h.db.QueryRow(userQuery, userID).Scan(&role)
	if err == nil && role == "superadmin" {
		return true // Superadmin has access to all elections
	}

	// Check if admin is assigned to this election
	var count int
	query := `SELECT COUNT(*) FROM election_admins WHERE user_id = ? AND election_id = ?`
	h.db.QueryRow(query, userID, electionID).Scan(&count)
	return count > 0
}

func (h *Handlers) getCandidatesByElection(electionID string) ([]models.Candidate, error) {
	query := `SELECT id, name, description, photo_url, order_num, created_at FROM candidates WHERE election_id = ? ORDER BY order_num`
	rows, err := h.db.Query(query, electionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []models.Candidate
	for rows.Next() {
		var candidate models.Candidate
		err := rows.Scan(
			&candidate.ID, &candidate.Name, &candidate.Description,
			&candidate.PhotoURL, &candidate.Order, &candidate.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

func (h *Handlers) getCandidateByID(id string) (*models.Candidate, error) {
	candidate := &models.Candidate{}
	query := `SELECT id, election_id, name, description, photo_url, order_num FROM candidates WHERE id = ?`

	err := h.db.QueryRow(query, id).Scan(
		&candidate.ID, &candidate.ElectionID, &candidate.Name,
		&candidate.Description, &candidate.PhotoURL, &candidate.Order,
	)

	return candidate, err
}

func (h *Handlers) getTokensByElection(electionID string) ([]models.VotingToken, error) {
	query := `SELECT id, token, is_used, used_at, created_at FROM voting_tokens WHERE election_id = ? ORDER BY created_at DESC`
	rows, err := h.db.Query(query, electionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []models.VotingToken
	for rows.Next() {
		var token models.VotingToken
		err := rows.Scan(&token.ID, &token.Token, &token.IsUsed, &token.UsedAt, &token.CreatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (h *Handlers) getVotesByElection(electionID string) ([]models.Vote, error) {
	query := `
		SELECT v.id, v.candidate_id, c.name, v.voted_at
		FROM votes v
		JOIN candidates c ON v.candidate_id = c.id
		WHERE v.election_id = ?
		ORDER BY v.voted_at DESC
	`
	rows, err := h.db.Query(query, electionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []models.Vote
	for rows.Next() {
		var vote models.Vote
		var candidateName string
		err := rows.Scan(&vote.ID, &vote.CandidateID, &candidateName, &vote.VotedAt)
		if err != nil {
			return nil, err
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

func (h *Handlers) getVoteCountsByElection(electionID string) ([]models.VoteCount, error) {
	query := `
		SELECT c.id, c.name, COUNT(v.id) as vote_count
		FROM candidates c
		LEFT JOIN votes v ON c.id = v.candidate_id
		WHERE c.election_id = ?
		GROUP BY c.id, c.name
		ORDER BY vote_count DESC, c.name
	`
	rows, err := h.db.Query(query, electionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var voteCounts []models.VoteCount
	for rows.Next() {
		var vc models.VoteCount
		err := rows.Scan(&vc.CandidateID, &vc.CandidateName, &vc.VoteCount)
		if err != nil {
			return nil, err
		}
		voteCounts = append(voteCounts, vc)
	}

	return voteCounts, nil
}

func (h *Handlers) getElectionStats(electionID string) (*models.ElectionStats, error) {
	stats := &models.ElectionStats{}

	h.db.QueryRow("SELECT COUNT(*) FROM voting_tokens WHERE election_id = ?", electionID).Scan(&stats.TotalTokens)
	h.db.QueryRow("SELECT COUNT(*) FROM voting_tokens WHERE election_id = ? AND is_used = TRUE", electionID).Scan(&stats.UsedTokens)
	h.db.QueryRow("SELECT COUNT(*) FROM votes WHERE election_id = ?", electionID).Scan(&stats.TotalVotes)
	h.db.QueryRow("SELECT COUNT(*) FROM candidates WHERE election_id = ?", electionID).Scan(&stats.TotalCandidates)

	return stats, nil
}

func (h *Handlers) getAdminStats(userID int) map[string]int {
	stats := make(map[string]int)

	var assignedElections, activeElections int
	h.db.QueryRow("SELECT COUNT(*) FROM election_admins WHERE user_id = ?", userID).Scan(&assignedElections)
	h.db.QueryRow(`
		SELECT COUNT(*) FROM elections e
		JOIN election_admins ea ON e.id = ea.election_id
		WHERE ea.user_id = ? AND e.status = 'active'
	`, userID).Scan(&activeElections)

	stats["assigned_elections"] = assignedElections
	stats["active_elections"] = activeElections

	return stats
}

func generateRandomToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
