package data

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

// Permissions is a slice of permission codes.
type Permissions []string

const (
	PermissionReadMovies  = "movies:read"
	PermissionWriteMovies = "movies:write"
)

// Include checks if s is in the permissions p.
func (p Permissions) Include(s string) bool {
	for _, p := range p {
		if s == p {
			return true
		}
	}
	return false
}

// PermissionModel is a wrapper of a DB connection pool.
type PermissionModel struct {
	DB *sql.DB
}

// GetAllForUser retrieves all permissions code from DB for the a user.
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	// Prepare the query.
	query := `
		SELECT permissions.code
		FROM permissions
		INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
		INNER JOIN users ON users_permissions.user_id = users.id
		WHERE users.id = $1`

	// Prepare the context.
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// Execute the query.
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Copy the permission codes from result rows.
	var permissions Permissions
	for rows.Next() {
		var permission string
		err = rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	// Return scan errors if there is any.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// AddForUser add permission codes to given user.
func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	// Prepare the query.
	query := `
		INSERT INTO users_permissions
		SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`

	// Prepare the context.
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// Execute the query
	args := []interface{}{userID, pq.Array(codes)}
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}
