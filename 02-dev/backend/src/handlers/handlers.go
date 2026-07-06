package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"climbing-gym-backend/src/config"
	"climbing-gym-backend/src/db"
	"climbing-gym-backend/src/models"
	"climbing-gym-backend/src/services"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Config *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{Config: cfg}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Auth handlers
type RequestCodeRequest struct {
	Phone string `json:"phone" binding:"required,startswith=+7"`
}

func (h *Handler) RequestCode(c *gin.Context) {
	var req RequestCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Неверный формат номера телефона",
		})
		return
	}

	result, err := services.RequestCode(req.Phone, &h.Config.SMS)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

type VerifyCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required,min=4,max=6"`
}

func (h *Handler) VerifyCode(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Неверный код или истёк",
		})
		return
	}

	result, err := services.VerifyCode(req.Phone, req.Code, &h.Config.JWT)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_code",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Reference data handlers
func (h *Handler) GetZones(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, description, max_capacity, duration_minutes FROM zones ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка БД"})
		return
	}
	defer rows.Close()

	var zones []models.Zone
	for rows.Next() {
		var z models.Zone
		err := rows.Scan(&z.ID, &z.Name, &z.Description, &z.MaxCapacity, &z.DurationMinutes)
		if err != nil {
			continue
		}
		zones = append(zones, z)
	}

	c.JSON(http.StatusOK, zones)
}

func (h *Handler) GetInstructors(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, photo_url, rating FROM instructors ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка БД"})
		return
	}
	defer rows.Close()

	var instructors []models.Instructor
	for rows.Next() {
		var i models.Instructor
		err := rows.Scan(&i.ID, &i.Name, &i.PhotoURL, &i.Rating)
		if err != nil {
			continue
		}
		instructors = append(instructors, i)
	}

	c.JSON(http.StatusOK, instructors)
}

func (h *Handler) GetEquipment(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, type, available_count, price_per_slot FROM equipment WHERE available_count > 0 ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка БД"})
		return
	}
	defer rows.Close()

	var equipment []models.Equipment
	for rows.Next() {
		var e models.Equipment
		err := rows.Scan(&e.ID, &e.Name, &e.Type, &e.AvailableCount, &e.PricePerSlot)
		if err != nil {
			continue
		}
		equipment = append(equipment, e)
	}

	c.JSON(http.StatusOK, equipment)
}

// Slot handlers
func (h *Handler) GetSlots(c *gin.Context) {
	zone := c.DefaultQuery("zone", "all")
	date := c.Query("date")
	dateFrom := c.DefaultQuery("dateFrom", time.Now().Format("2006-01-02"))
	dateTo := c.DefaultQuery("dateTo", time.Now().AddDate(0, 0, 7).Format("2006-01-02"))
	instructorID := c.Query("instructorId")
	onlyAvailable := c.DefaultQuery("onlyAvailable", "true")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit > 50 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	query := `
		SELECT s.id, s.start_time, s.total_places, s.free_places, s.price, s.status,
			z.id, z.name, z.description, z.max_capacity, z.duration_minutes,
			i.id, COALESCE(i.name, ''), COALESCE(i.photo_url, ''), COALESCE(i.rating, 0)
		FROM slots s
		JOIN zones z ON s.zone_id = z.id
		JOIN instructors i ON s.instructor_id = i.id
		WHERE s.status = 'available'
	`
	params := []interface{}{}
	paramIdx := 1

	if onlyAvailable == "true" {
		query += " AND s.free_places > 0"
	}

	query += " AND s.start_time >= $" + strconv.Itoa(paramIdx)
	params = append(params, dateFrom)
	paramIdx++
	query += " AND s.start_time <= $" + strconv.Itoa(paramIdx)
	params = append(params, dateTo)
	paramIdx++

	if zone != "all" {
		zoneMap := map[string]string{"boulder": "bouldering", "rope": "ropes"}
		zoneName := zoneMap[zone]
		if zoneName != "" {
			query += " AND z.name = $" + strconv.Itoa(paramIdx)
			params = append(params, zoneName)
			paramIdx++
		}
	}

	if date != "" {
		query += " AND s.start_time >= $" + strconv.Itoa(paramIdx)
		params = append(params, date)
		paramIdx++
		nextDay := time.Now()
		t, err := time.Parse("2006-01-02", date)
		if err == nil {
			nextDay = t.AddDate(0, 0, 1)
		}
		query += " AND s.start_time < $" + strconv.Itoa(paramIdx)
		params = append(params, nextDay.Format("2006-01-02"))
		paramIdx++
	}

	if instructorID != "" {
		query += " AND i.id = $" + strconv.Itoa(paramIdx)
		params = append(params, instructorID)
		paramIdx++
	}

	// Show all zones to all users

	query += " ORDER BY s.start_time ASC"
	query += " LIMIT $" + strconv.Itoa(paramIdx) + " OFFSET $" + strconv.Itoa(paramIdx+1)
	params = append(params, limit, (page-1)*limit)

	rows, err := db.Query(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка БД"})
		return
	}
	defer rows.Close()

	var slots []models.Slot
	for rows.Next() {
		var s models.Slot
		var zoneID, instructorID, maxParticipants, durationMinutes int
		var zoneName, zoneDesc, instructorName, instructorPhotoURL string
		var instructorRating float64

		err := rows.Scan(&s.ID, &s.StartTime, &s.TotalPlaces, &s.FreePlaces, &s.Price, &s.Status,
			&zoneID, &zoneName, &zoneDesc, &maxParticipants, &durationMinutes,
			&instructorID, &instructorName, &instructorPhotoURL, &instructorRating)
		if err != nil {
			fmt.Printf("[GetSlots] Scan error: %v\n", err)
			continue
		}

		s.Zone = models.Zone{ID: zoneID, Name: zoneName, Description: zoneDesc, MaxCapacity: maxParticipants, DurationMinutes: durationMinutes}
		s.Instructor = models.Instructor{ID: instructorID, Name: instructorName, PhotoURL: instructorPhotoURL, Rating: instructorRating}
		s.EndTime = s.StartTime.Add(time.Duration(durationMinutes) * time.Minute)
		if s.Status == "available" {
			s.Status = "available"
		}

		if strings.Contains(zoneName, "Болдеринг") {
			s.Level = "novice"
		} else {
			s.Level = "experienced"
		}

		slots = append(slots, s)
	}

	c.JSON(http.StatusOK, models.SlotsResponse{
		Data: slots,
		Meta: models.PaginationMeta{Page: page, Limit: limit, Total: len(slots), TotalPages: (len(slots) + limit - 1) / limit},
	})
}

func (h *Handler) GetSlotByID(c *gin.Context) {
	id := c.Param("id")

	var s models.Slot
	var zoneID, instructorID, maxParticipants, durationMinutes int
	var zoneName, zoneDesc, instructorName, instructorPhotoURL string
	var instructorRating float64

	err := db.QueryRow(`
		SELECT s.id, s.start_time, s.total_places, s.free_places, s.price, s.status,
			z.id, z.name, z.description, z.max_capacity, z.duration_minutes,
			i.id, i.name, i.photo_url, i.rating
		FROM slots s
		JOIN zones z ON s.zone_id = z.id
		JOIN instructors i ON s.instructor_id = i.id
		WHERE s.id = $1
	`, id).Scan(&s.ID, &s.StartTime, &s.TotalPlaces, &s.FreePlaces, &s.Price, &s.Status,
		&zoneID, &zoneName, &zoneDesc, &maxParticipants, &durationMinutes,
		&instructorID, &instructorName, &instructorPhotoURL, &instructorRating)

	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Слот не найден"})
		return
	}

	if s.Status == "cancelled" {
		c.JSON(http.StatusGone, models.ErrorResponse{Error: "slot_cancelled", Message: "Слот отменён скалодромом"})
		return
	}

	s.Zone = models.Zone{ID: zoneID, Name: zoneName, Description: zoneDesc, MaxCapacity: maxParticipants, DurationMinutes: durationMinutes}
	s.Instructor = models.Instructor{ID: instructorID, Name: instructorName, PhotoURL: instructorPhotoURL, Rating: instructorRating}
	s.EndTime = s.StartTime.Add(time.Duration(durationMinutes) * time.Minute)
	s.Status = "available"

	rows, _ := db.Query("SELECT id, name, type, available_count, price_per_slot FROM equipment WHERE available_count > 0")
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var e models.Equipment
			rows.Scan(&e.ID, &e.Name, &e.Type, &e.AvailableCount, &e.PricePerSlot)
			s.Equipment = append(s.Equipment, e)
		}
	}

	c.JSON(http.StatusOK, s)
}

type CreateSlotRequest struct {
	ZoneID      int    `json:"zoneId" binding:"required"`
	StartTime   string `json:"startTime" binding:"required"`
	TotalPlaces int    `json:"totalPlaces" binding:"required,min=1"`
	Price       int    `json:"price" binding:"required,min=0"`
}

func (h *Handler) CreateSlot(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "trainer" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "Доступ запрещён"})
		return
	}

	instructorID := c.GetInt("instructorID")
	if role == "trainer" && instructorID == 0 {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "Тренер не назначен"})
		return
	}

	var req CreateSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Ошибка валидации"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		startTime, err = time.Parse("2006-01-02T15:04", req.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Неверный формат даты"})
			return
		}
	}

	var zoneExists int
	err = db.QueryRow("SELECT 1 FROM zones WHERE id = $1", req.ZoneID).Scan(&zoneExists)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Зона не найдена"})
		return
	}

	actualInstructorID := instructorID
	if role == "admin" {
		actualInstructorID = 1
	}

	var slotID int
	err = db.QueryRow(`
		INSERT INTO slots (zone_id, instructor_id, start_time, total_places, free_places, price, status)
		VALUES ($1, $2, $3, $4, $4, $5, 'available')
		RETURNING id
	`, req.ZoneID, actualInstructorID, startTime, req.TotalPlaces, req.Price).Scan(&slotID)

	if err != nil {
		fmt.Printf("[CreateSlot] Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка создания слота"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": slotID, "message": "Слот создан"})
}

// Booking handlers
func (h *Handler) GetBookings(c *gin.Context) {
	clientID := c.GetInt("clientID")
	status := c.DefaultQuery("status", "all")
	upcoming := c.Query("upcoming")
	past := c.Query("past")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit > 50 {
		limit = 50
	}

	query := `
		SELECT b.id, b.client_id, b.slot_id, b.equipment_type, b.price, b.status, 
			COALESCE(b.cancellation_type, ''), COALESCE(b.cancellation_reason, ''), COALESCE(b.cancellation_fee, 0), b.created_at, b.updated_at,
			s.start_time, z.name as zone_name, i.name as instructor_name
		FROM bookings b
		JOIN slots s ON b.slot_id = s.id
		JOIN zones z ON s.zone_id = z.id
		JOIN instructors i ON s.instructor_id = i.id
		WHERE b.client_id = $1
	`
	params := []interface{}{clientID}
	paramIdx := 2

	if status != "all" {
		query += " AND b.status = $" + strconv.Itoa(paramIdx)
		params = append(params, status)
		paramIdx++
	}

	if upcoming == "true" {
		query += " AND s.start_time > NOW()"
	}

	if past == "true" {
		query += " AND s.start_time <= NOW()"
	}

	query += " ORDER BY s.start_time DESC"
	query += " LIMIT $" + strconv.Itoa(paramIdx) + " OFFSET $" + strconv.Itoa(paramIdx+1)
	params = append(params, limit, (page-1)*limit)

	rows, err := db.Query(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка БД"})
		return
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var b models.Booking
		var slotStartTime time.Time
		var zoneName, instructorName string
		var cancellationTypeStr, cancellationReasonStr string
		var cancellationFee float64

		err := rows.Scan(&b.ID, &b.ClientID, &b.SlotID, &b.EquipmentType, &b.Price, &b.Status,
			&cancellationTypeStr, &cancellationReasonStr, &cancellationFee, &b.CreatedAt, &b.UpdatedAt,
			&slotStartTime, &zoneName, &instructorName)
		if err != nil {
			fmt.Printf("[GetBookings] Scan error: %v\n", err)
			continue
		}

		b.CancellationType = &cancellationTypeStr
		b.CancellationReason = &cancellationReasonStr
		b.Slot = models.Slot{ID: b.SlotID, StartTime: slotStartTime}
		b.Slot.Zone = models.Zone{Name: zoneName}
		b.Slot.Instructor = models.Instructor{Name: instructorName}

		if b.Status == "cancelled_by_client_early" {
			b.CancellationPenalty = nil
		} else if b.Status == "cancelled_by_client_late" {
			b.CancellationPenalty = &cancellationFee
		}

		bookings = append(bookings, b)
	}

	c.JSON(http.StatusOK, models.BookingsResponse{
		Data: bookings,
		Meta: models.PaginationMeta{Page: page, Limit: limit, Total: len(bookings), TotalPages: (len(bookings) + limit - 1) / limit},
	})
}

type CreateBookingRequest struct {
	SlotID        int    `json:"slotId" binding:"required"`
	EquipmentType string `json:"equipmentType" binding:"required,oneof=own rental"`
	EquipmentID   *int   `json:"equipmentId"`
}

func (h *Handler) CreateBooking(c *gin.Context) {
	clientID := c.GetInt("clientID")
	role := c.GetString("role")

	if role != "client" && role != "" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "Только клиенты могут бронировать слоты"})
		return
	}

	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Ошибка валидации"})
		return
	}

	var slotPrice float64
	var freePlaces int
	var slotStatus string
	err := db.QueryRow("SELECT price, free_places, status FROM slots WHERE id = $1", req.SlotID).Scan(&slotPrice, &freePlaces, &slotStatus)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Слот не найден"})
		return
	}

	if slotStatus == "cancelled" {
		c.JSON(http.StatusGone, models.ErrorResponse{Error: "slot_cancelled", Message: "Слот отменён скалодромом"})
		return
	}

	if freePlaces <= 0 {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "slot_unavailable", Message: "Нет свободных мест"})
		return
	}

	var existingBooking int
	err = db.QueryRow("SELECT id FROM bookings WHERE client_id = $1 AND slot_id = $2 AND status NOT IN ('cancelled_by_client', 'cancelled_by_gym')", clientID, req.SlotID).Scan(&existingBooking)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка БД"})
		return
	}
	if existingBooking > 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "already_booked", Message: "Вы уже забронировали этот слот"})
		return
	}

	price := slotPrice
	if req.EquipmentType == "rental" && req.EquipmentID != nil {
		var rentalPrice float64
		err := db.QueryRow("SELECT price_per_slot FROM equipment WHERE id = $1", req.EquipmentID).Scan(&rentalPrice)
		if err != nil && err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка БД"})
			return
		}
		price += rentalPrice
	}

	confirmationDeadline := time.Now().Add(time.Duration(h.Config.Booking.ConfirmationDeadlineMinutes) * time.Minute)

	var bookingID int
	err = db.QueryRow(`
		INSERT INTO bookings (slot_id, client_id, equipment_type, status, price, confirmation_deadline)
		VALUES ($1, $2, $3, 'confirmed', $4, $5)
		RETURNING id
	`, req.SlotID, clientID, req.EquipmentType, price, confirmationDeadline).Scan(&bookingID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка создания брони"})
		return
	}

	db.Exec("UPDATE slots SET free_places = free_places - 1 WHERE id = $1", req.SlotID)

	c.JSON(http.StatusCreated, gin.H{
		"id":        bookingID,
		"status":    "confirmed",
		"price":     price,
		"expiresAt": confirmationDeadline,
	})
}

func (h *Handler) GetBookingByID(c *gin.Context) {
	clientID := c.GetInt("clientID")
	bookingID := c.Param("id")

	var b models.Booking
	var slotID int
	var slotStartTime time.Time
	var zoneID, instructorID, maxParticipants, durationMinutes int
	var zoneName, zoneDesc, instructorName, instructorPhotoURL string
	var instructorRating float64
	var cancellationFee float64

	err := db.QueryRow(`
		SELECT b.id, b.client_id, b.slot_id, b.equipment_type, b.price, b.status, 
			COALESCE(b.cancellation_type, ''), COALESCE(b.cancellation_reason, ''), b.created_at, b.updated_at,
			COALESCE(b.cancellation_fee, 0),
			s.id, s.start_time, s.total_places, s.free_places, s.price, s.status,
			z.id, z.name, z.description, z.max_capacity, z.duration_minutes,
			i.id, COALESCE(i.name, ''), COALESCE(i.photo_url, ''), COALESCE(i.rating, 0)
		FROM bookings b
		JOIN slots s ON b.slot_id = s.id
		JOIN zones z ON s.zone_id = z.id
		JOIN instructors i ON s.instructor_id = i.id
		WHERE b.id = $1 AND b.client_id = $2
	`, bookingID, clientID).Scan(
		&b.ID, &b.ClientID, &b.SlotID, &b.EquipmentType, &b.Price, &b.Status,
		&b.CancellationType, &b.CancellationReason, &b.CreatedAt, &b.UpdatedAt,
		&cancellationFee,
		&slotID, &slotStartTime, &b.Slot.TotalPlaces, &b.Slot.FreePlaces, &b.Slot.Price, &b.Slot.Status,
		&zoneID, &zoneName, &zoneDesc, &maxParticipants, &durationMinutes,
		&instructorID, &instructorName, &instructorPhotoURL, &instructorRating,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Бронирование не найдено"})
		return
	}
	if err != nil {
		fmt.Printf("[GetBookingByID] Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка БД"})
		return
	}

	b.Slot = models.Slot{
		ID:         slotID,
		StartTime:  slotStartTime,
		EndTime:    slotStartTime.Add(time.Duration(durationMinutes) * time.Minute),
		TotalPlaces: b.Slot.TotalPlaces,
		FreePlaces:  b.Slot.FreePlaces,
		Price:      b.Slot.Price,
		Status:     "available",
		Zone:       models.Zone{ID: zoneID, Name: zoneName, Description: zoneDesc, MaxCapacity: maxParticipants, DurationMinutes: durationMinutes},
		Instructor: models.Instructor{ID: instructorID, Name: instructorName, PhotoURL: instructorPhotoURL, Rating: instructorRating},
	}

	if b.Status == "cancelled_by_client_late" {
		b.CancellationPenalty = &cancellationFee
	}

	if b.EquipmentID != nil {
		var e models.Equipment
		err := db.QueryRow("SELECT id, name, type, available_count, price_per_slot FROM equipment WHERE id = $1", *b.EquipmentID).Scan(&e.ID, &e.Name, &e.Type, &e.AvailableCount, &e.PricePerSlot)
		if err == nil {
			b.Equipment = &e
		}
	}

	c.JSON(http.StatusOK, b)
}

func (h *Handler) CancelBooking(c *gin.Context) {
	clientID := c.GetInt("clientID")
	bookingID := c.Param("id")

	var slotID int
	var bookingStatus string
	var slotStartTime time.Time
	var slotPrice float64

	err := db.QueryRow(`
		SELECT b.slot_id, b.status, s.start_time, s.price
		FROM bookings b
		JOIN slots s ON b.slot_id = s.id
		WHERE b.id = $1 AND b.client_id = $2
	`, bookingID, clientID).Scan(&slotID, &bookingStatus, &slotStartTime, &slotPrice)

	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Бронирование не найдено"})
		return
	}

	if bookingStatus != "confirmed" && bookingStatus != "pending" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "already_cancelled", Message: "Бронирование уже отменено"})
		return
	}

	hoursUntilStart := time.Since(slotStartTime).Hours()
	var newStatus string
	var cancellationFee float64 = 0

	if hoursUntilStart >= float64(-h.Config.Booking.FreeCancellationHours) {
		newStatus = "cancelled_by_client_early"
	} else {
		newStatus = "cancelled_by_client_late"
		cancellationFee = slotPrice * float64(h.Config.Booking.LateCancellationPenaltyPercent) / 100
	}

	_, err = db.Exec(`
		UPDATE bookings 
		SET status = $1, cancellation_fee = $2, cancelled_at = NOW(), updated_at = NOW()
		WHERE id = $3
	`, newStatus, cancellationFee, bookingID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка отмены"})
		return
	}

	db.Exec("UPDATE slots SET free_places = free_places + 1 WHERE id = $1", slotID)

	c.JSON(http.StatusOK, gin.H{
		"id":                   bookingID,
		"status":               newStatus,
		"cancellationPenalty":  cancellationFee,
	})
}

// Profile handlers
func (h *Handler) GetProfile(c *gin.Context) {
	clientID := c.GetInt("clientID")
	fmt.Printf("[GetProfile] clientID: %d\n", clientID)

	client, err := services.GetClientProfile(clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Клиент не найден"})
		return
	}

	c.JSON(http.StatusOK, client)
}

type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	clientID := c.GetInt("clientID")

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Ошибка валидации"})
		return
	}

	if req.Level != "" && req.Level != "novice" && req.Level != "experienced" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Неверный уровень"})
		return
	}

	client, err := services.UpdateClientProfile(clientID, req.Name, req.Level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка обновления"})
		return
	}

	c.JSON(http.StatusOK, client)
}

func (h *Handler) GetAllUsers(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "Доступ запрещён"})
		return
	}

	rows, err := db.Query(`
		SELECT c.id, c.phone, COALESCE(c.name, ''), COALESCE(c.role, 'client'), 
			COALESCE(c.instructor_id, 0), c.created_at,
			COALESCE(i.name, '') as instructor_name
		FROM clients c
		LEFT JOIN instructors i ON c.instructor_id = i.id
		ORDER BY c.id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка БД"})
		return
	}
	defer rows.Close()

	type UserWithInstructor struct {
		models.Client
		InstructorName string `json:"instructorName"`
	}

	var users []UserWithInstructor
	for rows.Next() {
		var u UserWithInstructor
		var role string
		var instructorID int64
		var instructorName string
		err := rows.Scan(&u.ID, &u.Phone, &u.Name, &role, &instructorID, &u.CreatedAt, &instructorName)
		if err != nil {
			continue
		}
		u.Role = role
		if instructorID > 0 {
			instructorIDVal := int(instructorID)
			u.InstructorID = &instructorIDVal
		}
		u.InstructorName = instructorName
		users = append(users, u)
	}

	c.JSON(http.StatusOK, users)
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=client trainer admin"`
}

func (h *Handler) UpdateUserRole(c *gin.Context) {
	adminRole := c.GetString("role")
	if adminRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "Доступ запрещён"})
		return
	}

	userID := c.Param("id")
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Ошибка валидации"})
		return
	}

	_, err := db.Exec(`UPDATE clients SET role = $1, updated_at = NOW() WHERE id = $2`, req.Role, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка обновления"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Роль обновлена"})
}

type AssignTrainerRequest struct {
	Phone       string `json:"phone" binding:"required,startswith=+7"`
	InstructorID int   `json:"instructorId" binding:"required"`
}

func (h *Handler) AssignTrainer(c *gin.Context) {
	adminRole := c.GetString("role")
	if adminRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "Доступ запрещён"})
		return
	}

	var req AssignTrainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "validation_error", Message: "Ошибка валидации"})
		return
	}

	fmt.Printf("[AssignTrainer] phone=%s, instructorId=%d\n", req.Phone, req.InstructorID)

	var existingInstructor int
	err := db.QueryRow(`SELECT id FROM instructors WHERE id = $1`, req.InstructorID).Scan(&existingInstructor)
	if err != nil {
		fmt.Printf("[AssignTrainer] Instructor not found: %v\n", err)
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Тренер не найден"})
		return
	}

	_, err = db.Exec(`
		UPDATE clients 
		SET instructor_id = $1, role = 'trainer', updated_at = NOW() 
		WHERE phone = $2
	`, req.InstructorID, req.Phone)
	if err != nil {
		fmt.Printf("[AssignTrainer] Update error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка назначения тренера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Тренер назначен"})
}

func (h *Handler) GetAllBookings(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "trainer" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "Доступ запрещён"})
		return
	}

	rows, err := db.Query(`
		SELECT b.id, b.client_id, b.slot_id, b.equipment_type, b.price, b.status, 
			b.cancellation_type, b.cancellation_reason, b.created_at, b.updated_at,
			s.start_time, z.name as zone_name, i.name as instructor_name,
			c.phone as client_phone, COALESCE(c.name, '') as client_name
		FROM bookings b
		JOIN slots s ON b.slot_id = s.id
		JOIN zones z ON s.zone_id = z.id
		JOIN instructors i ON s.instructor_id = i.id
		JOIN clients c ON b.client_id = c.id
		WHERE b.status IN ('confirmed', 'pending')
		ORDER BY s.start_time DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка БД"})
		return
	}
	defer rows.Close()

	type BookingWithClient struct {
		models.Booking
		ClientPhone string `json:"clientPhone"`
		ClientName  string `json:"clientName"`
	}

	var bookings []BookingWithClient
	for rows.Next() {
		var b BookingWithClient
		var slotStartTime time.Time
		var zoneName, instructorName, clientPhone, clientName string
		var cancellationFee float64

		err := rows.Scan(&b.ID, &b.ClientID, &b.SlotID, &b.EquipmentType, &b.Price, &b.Status,
			&cancellationFee, &b.CancellationReason, &b.CreatedAt, &b.UpdatedAt, &b.UpdatedAt,
			&slotStartTime, &zoneName, &instructorName, &clientPhone, &clientName)
		if err != nil {
			continue
		}

		b.Slot = models.Slot{ID: b.SlotID, StartTime: slotStartTime}
		b.Slot.Zone = models.Zone{Name: zoneName}
		b.Slot.Instructor = models.Instructor{Name: instructorName}
		b.ClientPhone = clientPhone
		b.ClientName = clientName

		bookings = append(bookings, b)
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *Handler) AdminCancelBooking(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "trainer" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Message: "Доступ запрещён"})
		return
	}

	bookingID := c.Param("id")

	var slotID int
	err := db.QueryRow(`SELECT slot_id FROM bookings WHERE id = $1`, bookingID).Scan(&slotID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "not_found", Message: "Бронирование не найдено"})
		return
	}

	_, err = db.Exec(`
		UPDATE bookings 
		SET status = 'cancelled_by_gym', cancelled_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, bookingID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "internal_error", Message: "Ошибка отмены"})
		return
	}

	db.Exec("UPDATE slots SET free_places = free_places + 1 WHERE id = $1", slotID)

	c.JSON(http.StatusOK, gin.H{"message": "Бронирование отменено"})
}