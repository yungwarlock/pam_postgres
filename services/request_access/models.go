package requestaccess

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type requestAccessModel struct {
	DB *sql.DB
}

func (m *requestAccessModel) InitDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS access_requests (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		reason TEXT NOT NULL,
		status TEXT NOT NULL,
		permissions JSONB NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := m.DB.Exec(query)
	return err
}

func (m *requestAccessModel) CreateAccessRequest(ar *AccessRequest) error {
	query := `
	INSERT INTO access_requests (name, email, reason, status, permissions)
	VALUES ($1, $2, $3, $4, $5);`
	permissionsJSON, err := json.Marshal(ar.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %v", err)
	}
	_, err = m.DB.Exec(query, ar.Name, ar.Email, ar.Reason, ar.Status, string(permissionsJSON))
	return err
}

func (m *requestAccessModel) GetAllAccessRequests() (*[]AccessRequest, error) {
	query := `SELECT id, name, email, reason, status, permissions, created_at, updated_at FROM access_requests;`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query access requests: %v", err)
	}
	defer rows.Close()

	var accessRequests []AccessRequest
	for rows.Next() {
		var ar AccessRequest
		var permissionsJSON string
		if err := rows.Scan(&ar.ID, &ar.Name, &ar.Email, &ar.Reason, &ar.Status, &permissionsJSON, &ar.CreatedAt, &ar.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan access request: %v", err)
		}
		if err := json.Unmarshal([]byte(permissionsJSON), &ar.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %v", err)
		}
		accessRequests = append(accessRequests, ar)
	}
	return &accessRequests, nil
}

func (m *requestAccessModel) UpdateAccessRequestStatus(requestID string, status RequestStatus) error {
	query := `UPDATE access_requests SET status = $1 WHERE id = $2;`
	_, err := m.DB.Exec(query, status, requestID)
	return err
}
