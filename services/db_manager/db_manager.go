package dbmanager

import (
	"context"
	"database/sql"
	"fmt"
	"net/netip"
	"os"
	"time"

	"github.com/docker/docker/libnetwork/etchosts"
)

var (
	Debug = os.Getenv("DEBUG") == "1"
)

func GenerateTempUserWithPermissions(ctx context.Context, authDetails *PostgresAuthDetails, permissionSet *PermissionSet) (*PostgresAuthDetails, error) {
	timeout := 40 * time.Second

	dbStr := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres", authDetails.User, authDetails.Password, authDetails.Host, authDetails.Port)
	db, err := sql.Open("pgx", dbStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to admin database: %v", err)
	}
	defer db.Close()

	name, port, fullName := GenerateSubdomainAndPort()
	username, password, err := createTempUser(ctx, db, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary user: %v", err)
	}

	err = grantPermissions(ctx, authDetails, username, permissionSet)
	if err != nil {
		return nil, fmt.Errorf("failed to grant permissions: %v", err)
	}

	if Debug {
		fmt.Printf("Debug: Generated subdomain %s and port %s\n", fullName, port)
		err := etchosts.Add("/etc/hosts", []etchosts.Record{{
			Hosts: fullName,
			IP:    netip.MustParseAddr("127.0.0.1"),
		}})

		if err != nil {
			return nil, fmt.Errorf("failed to add /etc/hosts entry: %v", err)
		}
	}

	go CreateConnection(name, port, timeout)
	fmt.Printf("Access approved. Connect to %s on port %s\n", fullName, port)
	fmt.Printf("Login with username: %s and password: %s\n", username, password)

	return &PostgresAuthDetails{
		Host:     fullName,
		Port:     port,
		User:     username,
		Password: password,
	}, nil
}

func grantPermissions(ctx context.Context, authDetails *PostgresAuthDetails, userName string, permissionSet *PermissionSet) error {
	for database, dbPermissions := range *permissionSet {
		dbStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", authDetails.User, authDetails.Password, authDetails.Host, authDetails.Port, database)
		db, err := sql.Open("pgx", dbStr)
		if err != nil {
			return fmt.Errorf("failed to connect to database %s: %v", database, err)
		}

		defer db.Close()

		for table, tablePermissions := range dbPermissions {
			permissionsStr := ""
			for permission, allowed := range tablePermissions {
				if checkPermissionAllowed(permission) && allowed {
					permissionsStr += permission + ", "
				}
			}

			if len(permissionsStr) > 0 {
				// Remove trailing comma and space
				permissionsStr = permissionsStr[:len(permissionsStr)-2]
				grantQuery := "GRANT " + permissionsStr + " ON TABLE " + table + " TO " + userName + ";"
				_, err := db.ExecContext(ctx, grantQuery)
				if err != nil {
					return fmt.Errorf("failed to grant permissions on table %s in database %s: %v", table, database, err)
				}
			}

		}
	}
	return nil
}
