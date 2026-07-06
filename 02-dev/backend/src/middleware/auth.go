package middleware

import (
	"net/http"
	"strings"

	"climbing-gym-backend/src/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ClientID      int    `json:"clientId"`
	Phone         string `json:"phone"`
	Role          string `json:"role"`
	InstructorID  *int   `json:"instructorId,omitempty"`
	jwt.RegisteredClaims
}

func AuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Требуется авторизация",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Неверный формат токена",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Неверный токен",
			})
			c.Abort()
			return
		}

		c.Set("clientID", claims.ClientID)
		c.Set("phone", claims.Phone)
		c.Set("role", claims.Role)
		if claims.InstructorID != nil {
			c.Set("instructorID", *claims.InstructorID)
		}
		c.Next()
	}
}

func OptionalAuthMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err == nil && token.Valid {
			c.Set("clientID", claims.ClientID)
			c.Set("phone", claims.Phone)
			c.Set("role", claims.Role)
			if claims.InstructorID != nil {
				c.Set("instructorID", *claims.InstructorID)
			}
		}

		c.Next()
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			_ = c.Errors.Last()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Внутренняя ошибка сервера",
			})
		}
	}
}