package iamdb

import (
	"context"
	"database/sql"
	"fmt"
)

type RolePermissions struct {
	Role       string `json:"role,omitempty"`
	Permission string `json:"permission,omitempty"`
}

type IamDB struct {
	db *sql.DB
}

func NewIamDB(filename string) (*IamDB, error) {
	db, err := sql.Open("sqlite3", filename)

	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	return &IamDB{db: db}, nil
}

func query(db *sql.DB, sql string, bindVar string) ([]RolePermissions, error) {
	ctx := context.Background()
	rps := make([]RolePermissions, 0, 100)
	rows, err := db.QueryContext(ctx, sql, bindVar)
	if err != nil {
		return nil, fmt.Errorf("error inserting role: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rp RolePermissions
		err = rows.Scan(&rp.Role, &rp.Permission)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		// fmt.Printf("%s: %s\n", rp.Role, rp.Permission)
		rps = append(rps, rp)
	}

	return rps, nil
}

func (db *IamDB) QueryRoles(part string, wildcard bool) ([]RolePermissions, error) {

	bindVar := "roles/" + part
	if wildcard {
		bindVar = "roles/%" + part + "%"
	}

	return query(db.db, `
		SELECT r.name, p.permission
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE r.name LIKE ?`,
		bindVar)
}

func (db *IamDB) QueryPermissions(part string, wildcard bool) ([]RolePermissions, error) {

	bindVar := part
	queryString := ` 
		SELECT DISTINCT(r.name), ""
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE p.permission = ?`

	if wildcard {
		bindVar = "%" + part + "%"
		queryString = ` 
		SELECT r.name, p.permission
		FROM roles r
		JOIN role_permissions rp ON r.id = rp.role_id
		JOIN permissions p ON rp.permission_id = p.id
		WHERE p.permission LIKE ?`
	}

	return query(db.db, queryString, bindVar)
}

func (db *IamDB) Close() error {
	return db.db.Close()
}
