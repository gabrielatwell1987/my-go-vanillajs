package data

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"atwell.dev/reelingit/logger"
	"atwell.dev/reelingit/models"
	"github.com/go-webauthn/webauthn/webauthn"
)

// PasskeyRepository manages WebAuthn passkey data using a database.
type PasskeyRepository struct {
	db       *sql.DB                         // Database connection
	sessions map[string]webauthn.SessionData // In-memory session storage
	log      logger.Logger                   // Logger for debugging and errors
}

// NewPasskeyRepository initializes a new PasskeyRepository with a database connection.
func NewPasskeyRepository(db *sql.DB, log logger.Logger) *PasskeyRepository {
	return &PasskeyRepository{
		db:       db,
		sessions: make(map[string]webauthn.SessionData),
		log:      log,
	}
}

// GenSessionID generates a random session ID.
func (r *PasskeyRepository) GenSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetSession retrieves session data from the in-memory map.
func (r *PasskeyRepository) GetSession(token string) (webauthn.SessionData, bool) {
	r.log.Info(fmt.Sprintf("GetSession: %v", r.sessions[token]))
	val, ok := r.sessions[token]
	return val, ok
}

// SaveSession stores session data in the in-memory map.
func (r *PasskeyRepository) SaveSession(token string, data webauthn.SessionData) {
	r.log.Info(fmt.Sprintf("SaveSession: %s - %v", token, data))
	r.sessions[token] = data
}

// DeleteSession removes session data from the in-memory map.
func (r *PasskeyRepository) DeleteSession(token string) {
	r.log.Info(fmt.Sprintf("DeleteSession: %v", token))
	delete(r.sessions, token)
}

func (r *PasskeyRepository) GetUserByEmail(email string) (*models.PasskeyUser, error) {
	r.log.Info(fmt.Sprintf("Get User: %v", email))

	// Check if user exists by email
	var userID int
	err := r.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err == sql.ErrNoRows {
		r.log.Error("Failed to find new user", err)
		return nil, err
	} else if err != nil {
		r.log.Error("Failed to query user", err)
		return nil, err
	}

	// Fetch user credentials from passkeys table
	rows, err := r.db.Query("SELECT keys FROM passkeys WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to query passkeys", err)
		return nil, err
	}
	defer rows.Close()

	var credentials []webauthn.Credential
	for rows.Next() {
		var keys string
		if err := rows.Scan(&keys); err != nil {
			r.log.Error("Failed to scan passkey row", err)
			return nil, err
		}
		cred, err := deserializeCredential(keys)
		if err != nil {
			r.log.Error("Failed to deserialize credential", err)
			continue // Skip invalid credentials
		}
		credentials = append(credentials, cred)
	}

	// Construct and return PasskeyUser
	user := models.PasskeyUser{
		ID:          []byte(strconv.Itoa(userID)), // Convert int ID to byte slice
		Name:        email,
		DisplayName: email,
		Credentials: credentials,
	}
	return &user, nil
}

func (r *PasskeyRepository) GetUserByID(id int) (*models.PasskeyUser, error) {
	r.log.Info(fmt.Sprintf("Get User: %v", id))

	// Check if user exists by id
	var userID int
	err := r.db.QueryRow("SELECT id FROM users WHERE id = $1", id).Scan(&userID)
	if err == sql.ErrNoRows {
		r.log.Error("Failed to find new user", err)
		return nil, err
	} else if err != nil {
		r.log.Error("Failed to query user", err)
		return nil, err
	}

	// Fetch user credentials from passkeys table
	rows, err := r.db.Query("SELECT keys FROM passkeys WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to query passkeys", err)
		return nil, err
	}
	defer rows.Close()

	// Fetch the email of the user
	var email string
	err = r.db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Error("Failed to find user email", err)
			return nil, err
		}
		r.log.Error("Failed to query user email", err)
		return nil, err
	}

	var credentials []webauthn.Credential
	for rows.Next() {
		var keys string
		if err := rows.Scan(&keys); err != nil {
			r.log.Error("Failed to scan passkey row", err)
			return nil, err
		}
		cred, err := deserializeCredential(keys)
		if err != nil {
			r.log.Error("Failed to deserialize credential", err)
			continue // Skip invalid credentials
		}
		credentials = append(credentials, cred)
	}

	// Construct and return PasskeyUser
	user := models.PasskeyUser{
		ID:          []byte(strconv.Itoa(userID)), // Convert int ID to byte slice
		Name:        email,
		DisplayName: email,
		Credentials: credentials,
	}
	return &user, nil
}

// SaveUser updates the user's credentials in the database.
func (r *PasskeyRepository) SaveUser(user models.PasskeyUser) {
	r.log.Info(fmt.Sprintf("SaveUser: %v", user.WebAuthnName()))

	// Convert user ID from byte slice to integer
	userID, err := strconv.Atoi(string(user.ID))
	if err != nil {
		r.log.Error("Invalid user ID", err)
		return
	}

	// Insert new credentials
	for _, cred := range user.Credentials {
		keys, err := serializeCredential(cred)
		if err != nil {
			r.log.Error("Failed to serialize credential", err)
			continue
		}
		// Check if the key already exists in the database
		var exists bool
		err = r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM passkeys WHERE user_id = $1 AND keys = $2)", userID, keys).Scan(&exists)
		if err != nil {
			r.log.Error("Failed to check if passkey exists", err)
			continue
		}

		// Insert the key only if it does not already exist
		if !exists {
			_, err = r.db.Exec("INSERT INTO passkeys (user_id, keys) VALUES ($1, $2)", userID, keys)
			if err != nil {
				r.log.Error("Failed to insert passkey", err)
			}
		} else {
			r.log.Info(fmt.Sprintf("Passkey already exists for user_id: %d", userID))
		}
	}
}

// serializeCredential converts a WebAuthn credential to a JSON string.
func serializeCredential(cred webauthn.Credential) (string, error) {
	data, err := json.Marshal(cred)
	if err != nil {
		return "", fmt.Errorf("failed to marshal credential: %w", err)
	}
	return string(data), nil
}

// deserializeCredential converts a JSON string back to a WebAuthn credential.
func deserializeCredential(data string) (webauthn.Credential, error) {
	var cred webauthn.Credential
	err := json.Unmarshal([]byte(data), &cred)
	if err != nil {
		return webauthn.Credential{}, fmt.Errorf("failed to unmarshal credential: %w", err)
	}
	return cred, nil
}
