package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
}

func NewTokenManager(privateKeyPath, publicKeyPath string, accessTTL time.Duration) (*TokenManager, error) {
	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}
	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return nil, err
	}
	return &TokenManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		accessTTL:  accessTTL,
	}, nil
}

type accessClaims struct {
	Name string `json:"name"`
	Role Role   `json:"role"`
	jwt.RegisteredClaims
}

func (tm *TokenManager) IssueAccessToken(user *User) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(tm.accessTTL)
	claims := accessClaims{
		Name: user.Name,
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(tm.privateKey)
	if err != nil {
		return "", 0, err
	}
	return signed, int64(tm.accessTTL.Seconds()), nil
}

func (tm *TokenManager) ParseAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodRS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return tm.publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return &Claims{
		UserID: claims.Subject,
		Name:   claims.Name,
		Role:   claims.Role,
	}, nil
}

func (tm *TokenManager) NewRefreshToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(raw)
	hash := HashToken(token)
	return token, hash, nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("decode private key pem")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		parsed, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		rsaKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not rsa")
		}
		return rsaKey, nil
	}
	return key, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("decode public key pem")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	rsaKey, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not rsa")
	}
	return rsaKey, nil
}
