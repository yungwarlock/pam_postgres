package requestaccess

import dbmanager "pam_postgres/services/db_manager"

type RequestStatus string

const (
	StatusPending  RequestStatus = "pending"
	StatusApproved RequestStatus = "approved"
	StatusRejected RequestStatus = "rejected"
	StatusExpired  RequestStatus = "expired"
)

type AccessRequest struct {
	ID          int                           `json:"id"`
	Status      RequestStatus                 `json:"status"`
	CreatedAt   string                        `json:"created_at"`
	UpdatedAt   string                        `json:"updated_at"`
	Permissions dbmanager.PermissionSet       `json:"permissions"`
	AuthDetails dbmanager.PostgresAuthDetails `json:"auth_details"`
	Name        string                        `json:"name" validate:"required"`
	Reason      string                        `json:"reason" validate:"required"`
	Email       string                        `json:"email" validate:"required,email"`
}
