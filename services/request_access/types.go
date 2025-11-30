package requestaccess

type PermissionState map[string]map[string]bool

type RequestStatus string

const (
	StatusPending  RequestStatus = "pending"
	StatusApproved RequestStatus = "approved"
	StatusRejected RequestStatus = "rejected"
	StatusExpired  RequestStatus = "expired"
)

type AccessRequest struct {
	ID          int             `json:"id"`
	Status      RequestStatus   `json:"status"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
	Permissions PermissionState `json:"permissions"`
	Name        string          `json:"name" validate:"required"`
	Reason      string          `json:"reason" validate:"required"`
	Email       string          `json:"email" validate:"required,email"`
}
