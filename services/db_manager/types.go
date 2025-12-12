package dbmanager

type PostgresAuthDetails struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

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
