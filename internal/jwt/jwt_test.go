package jwt

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("SECRETKEY", "testsecret")
	os.Exit(m.Run())
}

func generateToken(t *testing.T, claims jwt.MapClaims, method jwt.SigningMethod, secret string) string {
	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString([]byte(secret))
	assert.NoError(t, err)
	return signed
}

func TestValidateToken_Valid(t *testing.T) {
	token := generateToken(t, jwt.MapClaims{
		"user_id": "123",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}, jwt.SigningMethodHS256, "testsecret")

	userID, err := ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "123", userID)
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	token := generateToken(t, jwt.MapClaims{
		"user_id": "123",
	}, jwt.SigningMethodHS256, "wrongsecret")

	_, err := ValidateToken(token)
	assert.EqualError(t, err, "invalid token")
}

func TestValidateToken_BadFormat(t *testing.T) {
	_, err := ValidateToken("this_is_not_a_token")
	assert.EqualError(t, err, "invalid token")
}

func TestValidateToken_InvalidMethod(t *testing.T) {

	tokenString := "invalid.token.signature"

	_, err := ValidateToken(tokenString)
	assert.EqualError(t, err, "invalid token")
}

func TestValidateToken_MissingUserID(t *testing.T) {
	token := generateToken(t, jwt.MapClaims{}, jwt.SigningMethodHS256, "testsecret")

	_, err := ValidateToken(token)
	assert.EqualError(t, err, "user_id missing")
}

func TestValidateToken_NonStringUserID(t *testing.T) {
	token := generateToken(t, jwt.MapClaims{
		"user_id": 123, // тип int, не string
	}, jwt.SigningMethodHS256, "testsecret")

	_, err := ValidateToken(token)
	assert.EqualError(t, err, "user_id missing")
}
