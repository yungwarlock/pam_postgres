package requestaccess

// type PermissionSet map[string]map[string]bool

// type PermissionSet struct {
// 	TableName  string `json:"table_name"`
// 	Select     bool   `json:"select"`
// 	Insert     bool   `json:"insert"`
// 	Update     bool   `json:"update"`
// 	Delete     bool   `json:"delete"`
// 	Truncate   bool   `json:"truncate"`
// 	References bool   `json:"references"`
// 	Trigger    bool   `json:"trigger"`
// }

type RequestStatus string

const (
	StatusPending  RequestStatus = "pending"
	StatusApproved RequestStatus = "approved"
	StatusRejected RequestStatus = "rejected"
	StatusExpired  RequestStatus = "expired"
)

var allowedPermissions = []string{
	"SELECT",
	"INSERT",
	"UPDATE",
	"DELETE",
	"TRUNCATE",
	"REFERENCES",
	"TRIGGER",
}

func checkPermissionAllowed(permission string) bool {
	for _, perm := range allowedPermissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// PermissionSet defines the structure for database permissions
// it follows this format
// database_name -> table_name -> permission -> allowed (bool)
type PermissionSet map[string]map[string]map[string]bool

type AccessRequest struct {
	ID          int           `json:"id"`
	Status      RequestStatus `json:"status"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
	Permissions PermissionSet `json:"permissions"`
	Name        string        `json:"name" validate:"required"`
	Reason      string        `json:"reason" validate:"required"`
	Email       string        `json:"email" validate:"required,email"`
}
