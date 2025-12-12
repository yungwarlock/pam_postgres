

// key is the permission name
type TablePermission = Record<string, boolean>;

// key is the table name
type DatabasePermissions = Record<string, TablePermission>;

// key is the database
export type PermissionSet = Record<string, DatabasePermissions>;

// key is the database name, value is array of table names
export type Databases = Record<string, string[]>; 