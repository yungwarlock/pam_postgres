package requestaccess

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	dbmanager "pam_postgres/services/db_manager"
)

type RequestAccessModel struct {
	rootUser      string
	rootPassword  string
	host          string
	port          string
	adminDatabase string
	DB            *sql.DB
}

func NewRequestAccessModel(rootUser, rootPassword, host, port, adminDatabase string) *RequestAccessModel {
	dbStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", rootUser, rootPassword, host, port, adminDatabase)
	db := setupDB(dbStr)

	return &RequestAccessModel{
		rootUser:      rootUser,
		rootPassword:  rootPassword,
		host:          host,
		port:          port,
		adminDatabase: adminDatabase,
		DB:            db,
	}
}

func (m *RequestAccessModel) InitDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS access_requests (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		reason TEXT NOT NULL,
		status TEXT NOT NULL,
		auth_details JSONB,
		permissions JSONB NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := m.DB.Exec(query)
	return err
}

func setupDB(url string) *sql.DB {
	db, err := sql.Open("pgx", url)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	return db
}

func (m *RequestAccessModel) CreateAccessRequest(ctx context.Context, ar *AccessRequest) error {
	query := `
	INSERT INTO access_requests (name, email, reason, status, auth_details, permissions)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id;`
	authDetailsJSON, err := json.Marshal(ar.AuthDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal auth details: %v", err)
	}
	permissionsJSON, err := json.Marshal(ar.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %v", err)
	}

	var lastInsertedID int
	err = m.DB.QueryRowContext(ctx, query, ar.Name, ar.Email, ar.Reason, ar.Status, string(authDetailsJSON), string(permissionsJSON)).Scan(&lastInsertedID)
	if err != nil {
		log.Fatalf("Error inserting access request: %v", err)
		return fmt.Errorf("failed to insert access request: %v", err)
	}

	ar.ID = lastInsertedID
	return nil
}

func (m *RequestAccessModel) GetAllAccessRequests(ctx context.Context) (*[]AccessRequest, error) {
	query := `SELECT id, name, email, reason, status, auth_details, permissions, created_at, updated_at FROM access_requests;`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query access requests: %v", err)
	}
	defer rows.Close()

	var accessRequests []AccessRequest
	for rows.Next() {
		var ar AccessRequest
		var authDetailsJSON string
		var permissionsJSON string
		if err := rows.Scan(&ar.ID, &ar.Name, &ar.Email, &ar.Reason, &ar.Status, &authDetailsJSON, &permissionsJSON, &ar.CreatedAt, &ar.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan access request: %v", err)
		}
		if err := json.Unmarshal([]byte(authDetailsJSON), &ar.AuthDetails); err != nil {
			return nil, fmt.Errorf("failed to unmarshal auth details: %v", err)
		}
		if err := json.Unmarshal([]byte(permissionsJSON), &ar.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %v", err)
		}
		accessRequests = append(accessRequests, ar)
	}
	return &accessRequests, nil
}

func (m *RequestAccessModel) GetAccessRequestByID(ctx context.Context, requestID string) (*AccessRequest, error) {
	query := `SELECT id, name, email, reason, status, auth_details, permissions,  created_at, updated_at FROM access_requests WHERE id = $1;`
	row := m.DB.QueryRowContext(ctx, query, requestID)

	var ar AccessRequest
	var authDetailsJSON string
	var permissionsJSON string
	if err := row.Scan(&ar.ID, &ar.Name, &ar.Email, &ar.Reason, &ar.Status, &authDetailsJSON, &permissionsJSON, &ar.CreatedAt, &ar.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan access request: %v", err)
	}
	if err := json.Unmarshal([]byte(authDetailsJSON), &ar.AuthDetails); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth details: %v", err)
	}
	if err := json.Unmarshal([]byte(permissionsJSON), &ar.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %v", err)
	}
	return &ar, nil
}

func (m *RequestAccessModel) UpdateAccessRequestStatus(ctx context.Context, requestID string, status RequestStatus) error {
	query := `UPDATE access_requests SET status = $1 WHERE id = $2;`
	_, err := m.DB.ExecContext(ctx, query, status, requestID)
	return err
}

func (m *RequestAccessModel) UpdateAccessRequestWithTempUser(ctx context.Context, requestID string, tempUserAuth *dbmanager.PostgresAuthDetails) error {
	query := `UPDATE access_requests SET auth_details = $1 WHERE id = $2;`
	authDetailsJSON, err := json.Marshal(tempUserAuth)
	if err != nil {
		return fmt.Errorf("failed to marshal auth details: %v", err)
	}
	_, err = m.DB.ExecContext(ctx, query, string(authDetailsJSON), requestID)
	return err
}

func (m *RequestAccessModel) GetAllDatabases(ctx context.Context) (*[]string, error) {
	query := `SELECT datname FROM pg_database WHERE datistemplate = false;`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %v", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, fmt.Errorf("failed to scan database name: %v", err)
		}
		databases = append(databases, dbName)
	}
	return &databases, nil
}

func (m *RequestAccessModel) GetAllTablesFromAllDatabases(ctx context.Context) (map[string][]string, error) {
	databasesQuery := `SELECT datname FROM pg_database WHERE datistemplate = false;`
	dbRows, err := m.DB.QueryContext(ctx, databasesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %v", err)
	}
	defer dbRows.Close()

	databaseTables := make(map[string][]string)

	for dbRows.Next() {
		var dbName string
		if err := dbRows.Scan(&dbName); err != nil {
			return nil, fmt.Errorf("failed to scan database name: %v", err)
		}

		dbStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", m.rootUser, m.rootPassword, m.host, m.port, dbName)
		db, err := sql.Open("pgx", dbStr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database %s: %v", dbName, err)
		}
		defer db.Close()

		tablesQuery := `SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema');`
		tableRows, err := db.QueryContext(ctx, tablesQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to query tables for database %s: %v", dbName, err)
		}

		var tables []string
		for tableRows.Next() {
			var tableName string
			if err := tableRows.Scan(&tableName); err != nil {
				tableRows.Close()
				return nil, fmt.Errorf("failed to scan table name for database %s: %v", dbName, err)
			}
			tables = append(tables, tableName)
		}
		tableRows.Close()

		if len(tables) > 0 {
			databaseTables[dbName] = tables
		}
	}

	return databaseTables, nil
}
