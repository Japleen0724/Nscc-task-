package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"attendance-system/internal/database"
	"attendance-system/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

func CreateEvent(c *gin.Context) {
	organizerID := c.GetInt64("organizer_id")

	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	radius := req.Radius
	if radius <= 0 {
		radius = 100.0 // default 100 meters
	}

	qrCode := fmt.Sprintf("EVENT-%d", time.Now().UnixNano()%100000000)

	res, err := database.DB.Exec(`
		INSERT INTO events (organizer_id, name, description, latitude, longitude, radius, qr_code, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, organizerID, req.Name, req.Description, req.Latitude, req.Longitude, radius, qrCode, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event: " + err.Error()})
		return
	}

	id, _ := res.LastInsertId()
	event := models.Event{
		ID:          id,
		OrganizerID: organizerID,
		Name:        req.Name,
		Description: req.Description,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Radius:      radius,
		QRCode:      qrCode,
		CreatedAt:   time.Now(),
	}

	c.JSON(http.StatusCreated, event)
}

func ListEvents(c *gin.Context) {
	organizerID := c.GetInt64("organizer_id")

	query := `
		SELECT e.id, e.organizer_id, e.name, e.description, e.latitude, e.longitude, e.radius, e.qr_code, e.created_at,
		       COUNT(a.id) as attendee_count
		FROM events e
		LEFT JOIN attendance a ON e.id = a.event_id
		WHERE e.organizer_id = ?
		GROUP BY e.id
		ORDER BY e.id DESC
	`

	rows, err := database.DB.Query(query, organizerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query events"})
		return
	}
	defer rows.Close()

	events := make([]models.Event, 0)
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(
			&e.ID, &e.OrganizerID, &e.Name, &e.Description,
			&e.Latitude, &e.Longitude, &e.Radius, &e.QRCode, &e.CreatedAt,
			&e.AttendeeCount,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan event"})
			return
		}
		events = append(events, e)
	}

	c.JSON(http.StatusOK, events)
}

func GetEvent(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var e models.Event
	query := `
		SELECT e.id, e.organizer_id, e.name, e.description, e.latitude, e.longitude, e.radius, e.qr_code, e.created_at,
		       COUNT(a.id) as attendee_count
		FROM events e
		LEFT JOIN attendance a ON e.id = a.event_id
		WHERE e.id = ?
		GROUP BY e.id
	`

	err = database.DB.QueryRow(query, id).Scan(
		&e.ID, &e.OrganizerID, &e.Name, &e.Description,
		&e.Latitude, &e.Longitude, &e.Radius, &e.QRCode, &e.CreatedAt,
		&e.AttendeeCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, e)
}

func LookupEventByQR(c *gin.Context) {
	qr := c.Query("qr")
	if qr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "QR code parameter is required"})
		return
	}

	var e models.Event
	query := `
		SELECT e.id, e.organizer_id, e.name, e.description, e.latitude, e.longitude, e.radius, e.qr_code, e.created_at,
		       COUNT(a.id) as attendee_count
		FROM events e
		LEFT JOIN attendance a ON e.id = a.event_id
		WHERE e.qr_code = ?
		GROUP BY e.id
	`

	err := database.DB.QueryRow(query, qr).Scan(
		&e.ID, &e.OrganizerID, &e.Name, &e.Description,
		&e.Latitude, &e.Longitude, &e.Radius, &e.QRCode, &e.CreatedAt,
		&e.AttendeeCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "No event found for this QR code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, e)
}

func GetEventQR(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var qrCode string
	err = database.DB.QueryRow("SELECT qr_code FROM events WHERE id = ?", id).Scan(&qrCode)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	png, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}

	c.Data(http.StatusOK, "image/png", png)
}
