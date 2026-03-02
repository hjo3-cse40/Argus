package auth
//consider using argon2id (alot more complex) but if we do oauth another implementation may be necessary.
import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

// hashPassword hashes a plaintext password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// comparePassword checks a plaintext password against a bcrypt hash
// Returns nil if they match, an error otherwise
func comparePassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
