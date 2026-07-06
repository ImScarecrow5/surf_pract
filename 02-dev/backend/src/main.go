package main

import (
	"fmt"
	"log"

	"climbing-gym-backend/src/config"
	"climbing-gym-backend/src/db"
	"climbing-gym-backend/src/handlers"
	"climbing-gym-backend/src/middleware"

	"github.com/gin-gonic/gin"
)

func runMigrations() {
	_, err := db.Exec("ALTER TABLE bookings ADD COLUMN IF NOT EXISTS confirmation_deadline TIMESTAMP")
	if err != nil {
		log.Printf("Migration warning: %v", err)
	} else {
		log.Println("Migration applied: confirmation_deadline column")
	}

	_, err = db.Exec("ALTER TABLE bookings ADD COLUMN IF NOT EXISTS cancellation_fee DECIMAL(10,2)")
	if err != nil {
		log.Printf("Migration warning: %v", err)
	} else {
		log.Println("Migration applied: cancellation_fee column")
	}
}

func main() {
	cfg := config.Load()

	if err := db.Connect(&cfg.DB); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	runMigrations()

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	r.Use(middleware.ErrorHandler())

	h := handlers.NewHandler(cfg)

	r.GET("/health", h.Health)

	v1 := r.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/request-code", h.RequestCode)
			auth.POST("/verify", h.VerifyCode)
		}

		reference := v1.Group("")
		{
			reference.GET("/zones", h.GetZones)
			reference.GET("/instructors", h.GetInstructors)
			reference.GET("/equipment", h.GetEquipment)
		}

		slots := v1.Group("/slots")
		{
			slots.GET("", middleware.OptionalAuthMiddleware(&cfg.JWT), h.GetSlots)
			slots.GET("/:id", h.GetSlotByID)
			slots.POST("", middleware.AuthMiddleware(&cfg.JWT), h.CreateSlot)
		}

		bookings := v1.Group("/bookings")
		bookings.Use(middleware.AuthMiddleware(&cfg.JWT))
		{
			bookings.GET("", h.GetBookings)
			bookings.GET("/:id", h.GetBookingByID)
			bookings.POST("", h.CreateBooking)
			bookings.PATCH("/:id/cancel", h.CancelBooking)
		}

		profile := v1.Group("/profile")
		profile.Use(middleware.AuthMiddleware(&cfg.JWT))
		{
			profile.GET("", h.GetProfile)
			profile.PATCH("", h.UpdateProfile)
		}

		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(&cfg.JWT))
		{
			admin.GET("/users", h.GetAllUsers)
			admin.PATCH("/users/:id/role", h.UpdateUserRole)
			admin.POST("/assign-trainer", h.AssignTrainer)
			admin.GET("/bookings", h.GetAllBookings)
			admin.PATCH("/bookings/:id/cancel", h.AdminCancelBooking)
		}
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server running on port %s (%s)", cfg.Port, cfg.NodeEnv)
	r.Run(addr)
}