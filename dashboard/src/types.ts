

// key is the permission name
type TablePermission = Record<string, boolean>;

// key is the table name
type DatabasePermissions = Record<string, TablePermission>;

// key is the database
export type PermissionSet = Record<string, DatabasePermissions>;

// key is the database name, value is array of table names
export type Databases = Record<string, string[]>;

export type RequestStatus = "waiting" | "approved" | "rejected";

export interface PostgresAuthDetails {
  user: string;
  password: string;
  host: string;
  port: number;
}

export interface AccessRequest {
  id: number;
  status: RequestStatus;
  created_at: string; // ISO timestamp
  updated_at: string; // ISO timestamp
  permissions: PermissionSet;
  auth_details: PostgresAuthDetails;
  name?: string;
  reason?: string;
  email?: string;
}