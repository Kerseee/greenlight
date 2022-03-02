package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"greenlight.kerseeehuang.com/internal/validator"
)

var (
	ErrDuplicateEmail   = errors.New("duplicate email") // "duplicate email"
	pqErrDuplicateEmail = `pq: duplicate key value violates unique constraint "users_email_key"`
)

// AnonymousUser represents an anonymous user. May be used in authentication middleware.
var AnonymousUser = &User{}

// User holds information of a user.
type User struct {
	ID        int64     `json:"id"`
	CreateAt  time.Time `json:"create_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// IsAnonymous returns true if the user u is an anonymous (inactivated) user.
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// UserModel is a wrapper of database connection pool.
type UserModel struct {
	DB *sql.DB
}

const passwordCost int = 12 // cost of hashing password

type password struct {
	plaintext *string // password that a client input
	hash      []byte  // hashed password
}

// Set hashes the plaintextPassword and stores it and hashed password into p.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), passwordCost)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

// Matches checkes if plaintextPassword match the hashed password of p.
// It returns true if matches, and false otherwise.
// Return non-nil error if there are problems during comparison.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// ValidateEmail validates email and stores error information into v.
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", validator.ErrMsgMustBeProvided)
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be valid email address")
}

// ValidatePlainPassword validates plain-text password and stores error information into v.
func ValidatePlainPassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", validator.ErrMsgMustBeProvided)
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes")
	v.Check(len(password) <= 72, "password", "must be at most 72 bytes")
}

// ValidateUserName validates name and stores error informaiton into v.
func ValidateUserName(v *validator.Validator, name string) {
	v.Check(name != "", "name", validator.ErrMsgMustBeProvided)
	v.Check(len(name) <= 500, "name", "must be at most 500 bytes")
}

// ValidateUser validates user and stores error information into v.
func ValidateUser(v *validator.Validator, user *User) {
	ValidateUserName(v, user.Name)
	ValidateEmail(v, user.Email)
	if user.Password.plaintext != nil {
		ValidatePlainPassword(v, *user.Password.plaintext)
	}
	if user.Password.hash == nil {
		panic("missing hash for user password")
	}
}

// Insert inserts a user into users table in the DB.
// If the user's email has used by other users, return data.ErrDuplicateEmail.
func (m UserModel) Insert(user *User) error {
	// Prepare query and arguments.
	query := `
		INSERT INTO users (name, email, password_hash, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, create_at, version`
	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}

	// Prepare context for executing db query.
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// Execute query.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreateAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == pqErrDuplicateEmail:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// GetByEmail return a user with given email.
func (m UserModel) GetByEmail(email string) (*User, error) {
	// Prepare a query.
	query := `
		SELECT id, create_at, name, password_hash, activated, version
		FROM users
		WHERE email = $1`

	// Prepare a context for executing query
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// Execute the query and store results into user.
	var user User
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreateAt,
		&user.Name,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Update updates the record of user in the DB.
func (m UserModel) Update(user *User) error {
	// Prepare the query and arguments.
	query := `
		UPDATE users
		SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`
	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	// Prepare a context for executing the query.
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// Execute the query.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == pqErrDuplicateEmail:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

// GetForToken return a users from DB with given scope and tokenPlaintext.
// If no matching token is found in DB, return nil User and data.ErrRecordNotFound.
func (m UserModel) GetForToken(scope string, tokenPlaintext string) (*User, error) {
	// Prepare the query.
	query := `
		SELECT users.id, users.create_at, users.name, users.email, users.password_hash, users.activated, users.version
		FROM users
		INNER JOIN tokens ON users.id = tokens.user_id
		WHERE tokens.hash = $1
		AND tokens.scope = $2
		AND tokens.expiry > $3`

	// Hash the tokenPlaintext.
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	// Prepare the arguments.
	args := []interface{}{tokenHash[:], scope, time.Now()}

	// Execute the query
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreateAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
