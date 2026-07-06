package models

import (
	"time"
)

type Client struct {
	ID         int       `json:"id"`
	Phone      string    `json:"phone"`
	Name       string    `json:"name"`
	ClientType string    `json:"level"`
	Role       string    `json:"role"`
	InstructorID *int    `json:"instructorId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Zone struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	MaxCapacity     int    `json:"maxCapacity"`
	DurationMinutes int    `json:"durationMinutes"`
}

type Instructor struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	PhotoURL string  `json:"photoUrl"`
	Rating   float64 `json:"rating"`
}

type Equipment struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	AvailableCount int     `json:"availableCount"`
	PricePerSlot   float64 `json:"pricePerSlot"`
}

type Slot struct {
	ID           int       `json:"id"`
	ZoneID       int       `json:"zoneId"`
	Zone         Zone      `json:"zone"`
	InstructorID int       `json:"instructorId"`
	Instructor   Instructor `json:"instructor"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	TotalPlaces  int       `json:"totalPlaces"`
	FreePlaces   int       `json:"freePlaces"`
	Price        float64   `json:"price"`
	Status       string    `json:"status"`
	Level        string    `json:"level,omitempty"`
	Equipment    []Equipment `json:"equipment,omitempty"`
}

type Booking struct {
	ID                int        `json:"id"`
	ClientID          int        `json:"clientId"`
	SlotID            int        `json:"slotId"`
	Slot              Slot       `json:"slot"`
	EquipmentType     string     `json:"equipmentType"`
	EquipmentID       *int       `json:"equipmentId,omitempty"`
	Equipment         *Equipment `json:"equipment,omitempty"`
	Price             float64    `json:"price"`
	Status            string     `json:"status"`
	CancellationType  *string    `json:"cancellationType,omitempty"`
	CancellationReason *string   `json:"cancellationReason,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	CancellationPenalty *float64  `json:"cancellationPenalty,omitempty"`
}

type AuthResponse struct {
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken"`
	ExpiresIn    int     `json:"expiresIn"`
	Client       *Client `json:"client"`
}

type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type SlotsResponse struct {
	Data []Slot          `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type BookingsResponse struct {
	Data []Booking       `json:"data"`
	Meta PaginationMeta `json:"meta"`
}