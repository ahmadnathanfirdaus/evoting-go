package middleware

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"evoting-app/internal/models"

	"github.com/gorilla/sessions"
)

type contextKey string

const UserContextKey contextKey = "user"

// Global session store - should be initialized from main
var globalStore sessions.Store

func SetSessionStore(store sessions.Store) {
	globalStore = store
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("RequireAuth middleware called for path: %s", r.URL.Path)
		session, err := globalStore.Get(r, "session")
		if err != nil {
			log.Printf("Error getting session: %v", err)
		}

		userID, ok := session.Values["user_id"]
		log.Printf("Session values: %+v, userID ok: %v", session.Values, ok)
		if !ok {
			log.Printf("No user_id in session, redirecting to login")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get user from database (you'll need to pass db connection)
		// For now, we'll create a simple user object
		user := &models.User{
			ID:   userID.(int),
			Role: session.Values["role"].(string),
		}
		log.Printf("User authenticated: %+v", user)

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil || user.Role != "superadmin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil || (user.Role != "admin" && user.Role != "superadmin") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// AuthService handles authentication logic
type AuthService struct {
	db *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

func (a *AuthService) GetUserByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, password, role, created_at, updated_at FROM users WHERE id = ?`

	err := a.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a *AuthService) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, password, role, created_at, updated_at FROM users WHERE username = ?`

	err := a.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}
