package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	JWTSecret = os.Getenv("JWT_SECRET")
	Debug     = os.Getenv("DEBUG") != ""
)

type contextKey string

const UserContextKey contextKey = "user"

type UserContext struct {
	ID       string `json:"id"`
	Exp      int64  `json:"exp"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// CSRFMiddleware generates and adds CSRF token to response headers
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := generateCSRFToken()

		SetCSRFCookie(w, token) // Set the token in an HTTP-only cookie

		next.ServeHTTP(w, r)
	})
}

func SetCSRFCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		HttpOnly: true,
		Secure:   !Debug, // Use secure cookies in production
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   3600, // Matches CSRF token validity (1 hour)
	}
	http.SetCookie(w, cookie)
}

func ClearCSRFCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		HttpOnly: true,
		Secure:   !Debug,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1, // Expire immediately
	}
	http.SetCookie(w, cookie)
}

// CSRFValidationMiddleware validates CSRF tokens for API calls
func CSRFValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip validation for GET requests
		if r.Method == "GET" {
			next.ServeHTTP(w, r)
			return
		}

		// Get token from HTTP-only cookie
		cookie, err := r.Cookie("csrf_token")
		if err != nil {
			// No CSRF cookie found, forbidden
			http.Error(w, "CSRF token required", http.StatusForbidden)
			return
		}
		cookieToken := cookie.Value

		// Get token from header (for comparison if client sends it, or for SPA)
		headerToken := r.Header.Get("X-CSRF-Token")

		// For double-submit cookie pattern, both tokens must be present and match
		// And both need to be validated
		if cookieToken == "" || headerToken == "" || cookieToken != headerToken {
			ClearCSRFCookie(w) // Clear invalid/missing cookie
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		if !validateCSRFToken(cookieToken) { // Validate the token from the cookie (or header, since they should match)
			ClearCSRFCookie(w) // Clear invalid cookie
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		// If validation passes, regenerate and set a new CSRF token for the next request
		newToken := generateCSRFToken()
		SetCSRFCookie(w, newToken)

		next.ServeHTTP(w, r)
	})
}

// SetAuthCookie sets an HTTP-only cookie with the JWT token
func SetAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		Secure:   !Debug, // Use secure cookies in production
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
	}
	http.SetCookie(w, cookie)
}

// ClearAuthCookie removes the authentication cookie
func ClearAuthCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		HttpOnly: true,
		Secure:   !Debug,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1, // Expire immediately
	}
	http.SetCookie(w, cookie)
}

// ValidateAuth validates JWT authentication from HTTP-only cookie and returns user context
func ValidateAuth(w http.ResponseWriter, r *http.Request) (*UserContext, error) {
	// Get token from HTTP-only cookie
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return nil, fmt.Errorf("authentication cookie not found")
	}

	token := cookie.Value
	if token == "" {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return nil, fmt.Errorf("empty authentication token")
	}

	user, err := validateJWTToken(token)
	if err != nil {
		if Debug {
			log.Printf("JWT validation error: %v", err)
		}
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return nil, fmt.Errorf("invalid or expired token")
	}

	return user, nil
}

// validateJWTToken validates a JWT token and returns user information
func validateJWTToken(token string) (*UserContext, error) {
	var user UserContext

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode header and payload
	payload, err := base64.URLEncoding.DecodeString(addPadding(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding")
	}

	signature, err := base64.URLEncoding.DecodeString(addPadding(parts[2]))
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding")
	}

	// Verify signature
	expectedSignature := generateJWTSignature(parts[0] + "." + parts[1])
	if !hmac.Equal(signature, expectedSignature) {
		return nil, fmt.Errorf("invalid signature")
	}

	if err := json.Unmarshal(payload, &user); err != nil {
		return nil, fmt.Errorf("invalid payload")
	}

	// Check expiration
	if time.Now().Unix() > user.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &user, nil
}

// GenerateJWTToken creates a JWT token for the user
func GenerateJWTToken(id, email string) (string, error) {
	user := UserContext{
		ID:    id,
		Email: email,
		Exp:   time.Now().Add(24 * time.Hour).Unix(),
	}

	// Create header
	header := map[string]string{
		"typ": "JWT",
		"alg": "HS256",
	}
	headerJSON, _ := json.Marshal(header)
	headerEncoded := base64.URLEncoding.EncodeToString(headerJSON)

	// Create payload
	payloadJSON, err := json.Marshal(user)
	if err != nil {
		return "", err
	}
	payloadEncoded := base64.URLEncoding.EncodeToString(payloadJSON)

	// Create signature
	message := headerEncoded + "." + payloadEncoded
	signature := generateJWTSignature(message)
	signatureEncoded := base64.URLEncoding.EncodeToString(signature)

	return message + "." + signatureEncoded, nil
}

// generateJWTSignature creates HMAC signature for JWT
func generateJWTSignature(message string) []byte {
	h := hmac.New(sha256.New, []byte(JWTSecret)) // Use dedicated JWT secret
	h.Write([]byte(message))
	return h.Sum(nil)
}

// addPadding adds padding to base64 string if needed
func addPadding(s string) string {
	switch len(s) % 4 {
	case 2:
		return s + "=="
	case 3:
		return s + "="
	}
	return s
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(r *http.Request) (*UserContext, bool) {
	user, ok := r.Context().Value(UserContextKey).(*UserContext)
	return user, ok
}

// generateCSRFToken creates an HMAC-based CSRF token with timestamp and nonce
func generateCSRFToken() string {
	timestamp := time.Now().Unix()
	nonce := generateNonce()
	message := fmt.Sprintf("%d.%s", timestamp, nonce)

	h := hmac.New(sha256.New, []byte(JWTSecret))
	h.Write([]byte(message))
	signature := h.Sum(nil)

	token := fmt.Sprintf("%s.%s", message, base64.URLEncoding.EncodeToString(signature))
	return base64.URLEncoding.EncodeToString([]byte(token))
}

// validateCSRFToken validates an HMAC-based CSRF token
func validateCSRFToken(token string) bool {
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false
	}

	parts := strings.Split(string(decoded), ".")
	if len(parts) != 3 {
		return false
	}

	timestampStr := parts[0]
	nonce := parts[1]
	signature, err := base64.URLEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	message := fmt.Sprintf("%s.%s", timestampStr, nonce)

	// Validate timestamp (token expires after 1 hour)
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix()-timestamp > 3600 {
		return false
	}

	// Validate HMAC signature
	h := hmac.New(sha256.New, []byte(JWTSecret))
	h.Write([]byte(message))
	expectedSignature := h.Sum(nil)

	return hmac.Equal(signature, expectedSignature)
}

// generateNonce creates a random nonce for unique token generation
func generateNonce() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based nonce if random fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}
