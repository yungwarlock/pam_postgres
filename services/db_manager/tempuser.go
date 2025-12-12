package dbmanager

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

func createTempUser(ctx context.Context, db *sql.DB, timeout time.Duration) (string, string, error) {
	username := generateTempUsername()
	password := generateTempPassword()

	// convert timeout to SQL format and create user
	timeoutStr := time.Now().Add(timeout).Format("2006-01-02 15:04:05")

	query := fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s' VALID UNTIL '%s'", username, password, timeoutStr)
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return "", "", err
	}

	return username, password, nil
}

func revokeTempUser(ctx context.Context, db *sql.DB, username string) error {
	query := fmt.Sprintf("DROP USER IF EXISTS %s", username)
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func generateTempUsername() string {
	// Generate 8 random bytes and encode as hex
	bytes := make([]byte, 8)
	rand.Read(bytes)

	return "temp_" + hex.EncodeToString(bytes)
}

func generateTempPassword() string {
	// Character sets for password generation
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits    = "0123456789"
	)

	charset := lowercase + uppercase + digits
	password := make([]byte, 16)

	for i := range password {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[num.Int64()]
	}

	return string(password)
}
