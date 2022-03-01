package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"greenlight.kerseeehuang.com/internal/validator"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

const (
	TokenExpireTimeActivation     = time.Hour * 24 * 3
	TokenExpireTimeAuthentication = time.Hour * 24
)

// Token holds attributes of a token stored in DB.
type Token struct {
	Plaintext string    `json:"token"` // Token string being sent to customer
	Hash      []byte    `json:"-"`     // Hashed token being stored in DB
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// generateToken generates a token based on given userID, expire time (ttl) and used scope.
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// Create a token struct.
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Generate random bytes.
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Store random bytes into token.
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Hash the plain text.
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

// validateTokenPlaintext validates the tokenPlaintext and store error messages into v.
// It check that tokenPlaintext is not empty and is exactly 26 bytes long.
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", validator.ErrMsgMustBeProvided)
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 tyes long")
}

// TokenModel is a wrapper of DB connection pool.
type TokenModel struct {
	DB *sql.DB
}

// Insert inserts token into DB.
func (m TokenModel) Insert(token *Token) error {
	// Prepare a query and arguments.
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)`

	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	// Prepare context for executing the query
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// Execute the query.
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// New creates a new token based on given userID, expire time (ttl) and used scope.
// New stores the new token into DB and also returns it.
// If errors happen, return nil token and error.
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// Create a new token.
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	// Insert the new token into DB.
	err = m.Insert(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// DeleteAllForUser deletes all tokens for the given user and specific scope.
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
		DELETE FROM tokens
		WHERE scope = $1 AND user_id = $2`

	args := []interface{}{scope, userID}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}
