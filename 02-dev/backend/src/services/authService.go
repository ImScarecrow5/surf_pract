package services

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"climbing-gym-backend/src/config"
	"climbing-gym-backend/src/db"
	"climbing-gym-backend/src/middleware"
	"climbing-gym-backend/src/models"

	"github.com/golang-jwt/jwt/v5"
)

var verificationCodes = make(map[string]*verificationCode)

type verificationCode struct {
	code     string
	expiresAt time.Time
}

func generateCode() string {
	code := rand.Intn(9000) + 1000
	return fmt.Sprintf("%04d", code)
}

func RequestCode(phone string, cfg *config.SMSConfig) (map[string]interface{}, error) {
	code := generateCode()
	expiresAt := time.Now().Add(time.Duration(cfg.CodeExpires) * time.Second)

	verificationCodes[phone] = &verificationCode{
		code:     code,
		expiresAt: expiresAt,
	}

	fmt.Printf("[SMS Mock] Code for %s: %s\n", phone, code)

	return map[string]interface{}{
		"success":   true,
		"message":   "Код подтверждения отправлен",
		"expiresIn": cfg.CodeExpires,
		"code":      code, // For dev only
	}, nil
}

func VerifyCode(phone, code string, cfg *config.JWTConfig) (*models.AuthResponse, error) {
	// Dev mode: accept 1234 for any phone
	if code != "1234" {
		stored, exists := verificationCodes[phone]
		if !exists {
			return nil, fmt.Errorf("код не найден или истёк")
		}

		if time.Now().After(stored.expiresAt) {
			delete(verificationCodes, phone)
			return nil, fmt.Errorf("код истёк")
		}

		if stored.code != code {
			return nil, fmt.Errorf("неверный код")
		}

		delete(verificationCodes, phone)
	}

	var client models.Client
	var name sql.NullString
	var role sql.NullString
	var instructorID sql.NullInt64
	err := db.QueryRow(`
		SELECT id, phone, COALESCE(name, ''), COALESCE(role, 'client'), COALESCE(instructor_id, 0), created_at
		FROM clients WHERE phone = $1
	`, phone).Scan(&client.ID, &client.Phone, &name, &role, &instructorID, &client.CreatedAt)
	client.Name = name.String
	client.Role = role.String
	if instructorID.Int64 > 0 {
		instructorIDVal := int(instructorID.Int64)
		client.InstructorID = &instructorIDVal
	}

	if err == sql.ErrNoRows {
		err = db.QueryRow(`
			INSERT INTO clients (phone, client_type, role) VALUES ($1, 'novice', 'client')
			RETURNING id, phone, COALESCE(name, ''), COALESCE(role, 'client'), COALESCE(instructor_id, 0), created_at
		`, phone).Scan(&client.ID, &client.Phone, &name, &role, &instructorID, &client.CreatedAt)
		client.Name = name.String
		client.Role = role.String
		if instructorID.Int64 > 0 {
			instructorIDVal := int(instructorID.Int64)
			client.InstructorID = &instructorIDVal
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	expiresInDuration, _ := time.ParseDuration(cfg.ExpiresIn)
	expiresIn := int(expiresInDuration.Seconds())

	claims := &middleware.Claims{
		ClientID:     client.ID,
		Phone:        client.Phone,
		Role:         client.Role,
		InstructorID: client.InstructorID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresInDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	refreshExpiresIn, _ := time.ParseDuration(cfg.RefreshExpiresIn)
	refreshClaims := &middleware.Claims{
		ClientID:     client.ID,
		Phone:        client.Phone,
		Role:         client.Role,
		InstructorID: client.InstructorID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(cfg.RefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresIn:    expiresIn,
		Client:       &client,
	}, nil
}

func GetClientProfile(clientID int) (*models.Client, error) {
	var client models.Client
	var role sql.NullString
	var clientType sql.NullString
	var instructorID sql.NullInt64
	err := db.QueryRow(`
		SELECT id, phone, COALESCE(name, ''), COALESCE(role, 'client'), COALESCE(client_type, 'novice'), COALESCE(instructor_id, 0), created_at
		FROM clients WHERE id = $1
	`, clientID).Scan(&client.ID, &client.Phone, &client.Name, &role, &clientType, &instructorID, &client.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("client not found")
	}

	client.Role = role.String
	client.ClientType = clientType.String
	if instructorID.Int64 > 0 {
		instructorIDVal := int(instructorID.Int64)
		client.InstructorID = &instructorIDVal
	}

	return &client, nil
}

func UpdateClientProfile(clientID int, name, level string) (*models.Client, error) {
	if name != "" {
		_, err := db.Exec(`UPDATE clients SET name = $1, updated_at = NOW() WHERE id = $2`, name, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to update name: %w", err)
		}
	}

	if level != "" {
		_, err := db.Exec(`UPDATE clients SET client_type = $1, updated_at = NOW() WHERE id = $2`, level, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to update level: %w", err)
		}
	}

	return GetClientProfile(clientID)
}