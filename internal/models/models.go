package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password"`
	Role      string    `json:"role" db:"role"` // "superadmin" or "admin"
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Election struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	StartDate   time.Time `json:"start_date" db:"start_date"`
	EndDate     time.Time `json:"end_date" db:"end_date"`
	Status      string    `json:"status" db:"status"` // "draft", "active", "completed"
	CreatedBy   int       `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Candidate struct {
	ID          int    `json:"id" db:"id"`
	ElectionID  int    `json:"election_id" db:"election_id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	PhotoURL    string `json:"photo_url" db:"photo_url"`
	Order       int    `json:"order" db:"order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type VotingToken struct {
	ID         int       `json:"id" db:"id"`
	ElectionID int       `json:"election_id" db:"election_id"`
	Token      string    `json:"token" db:"token"`
	IsUsed     bool      `json:"is_used" db:"is_used"`
	UsedAt     *time.Time `json:"used_at" db:"used_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type Vote struct {
	ID          int       `json:"id" db:"id"`
	ElectionID  int       `json:"election_id" db:"election_id"`
	CandidateID int       `json:"candidate_id" db:"candidate_id"`
	TokenID     int       `json:"token_id" db:"token_id"`
	VotedAt     time.Time `json:"voted_at" db:"voted_at"`
}

type ElectionAdmin struct {
	ID         int       `json:"id" db:"id"`
	ElectionID int       `json:"election_id" db:"election_id"`
	UserID     int       `json:"user_id" db:"user_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
}

// View models for reports
type VoteCount struct {
	CandidateID   int    `json:"candidate_id" db:"candidate_id"`
	CandidateName string `json:"candidate_name" db:"candidate_name"`
	VoteCount     int    `json:"vote_count" db:"vote_count"`
}

type ElectionStats struct {
	TotalTokens     int `json:"total_tokens" db:"total_tokens"`
	UsedTokens      int `json:"used_tokens" db:"used_tokens"`
	TotalVotes      int `json:"total_votes" db:"total_votes"`
	TotalCandidates int `json:"total_candidates" db:"total_candidates"`
}
