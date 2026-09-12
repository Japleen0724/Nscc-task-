package handlers

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"attendance-system/internal/database"
	"attendance-system/internal/models"
	"attendance-system/internal/utils"
	"github.com/gin-gonic/gin"
)

func CheckIn(c *gin.Context) {
	var req models.CheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: attendee_name, latitude, and longitude are required"})
		return
	}

	trimmedName := strings.TrimSpace(req.AttendeeName)
	if trimmedName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Attendee name cannot be empty"})
		return
	}

	var event models.Event
	var err error

	if req.EventID > 0 {
		err = database.DB.QueryRow(`
			SELECT id, name, latitude, longitude, radius FROM events WHERE id = ?
		`, req.EventID).Scan(&event.ID, &event.Name, &event.Latitude, &event.Longitude, &event.Radius)
	} else if req.QRCode != "" {
		err = database.DB.QueryRow(`
			SELECT id, name, latitude, longitude, radius FROM events WHERE qr_code = ?
		`, strings.TrimSpace(req.QRCode)).Scan(&event.ID, &event.Name, &event.Latitude, &event.Longitude, &event.Radius)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either event_id or qr_code must be provided"})
		return
	}

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found. Please verify the event code or QR"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	// 1. Prevent duplicate check-ins
	var existingID int64
	err = database.DB.QueryRow(`
		SELECT id FROM attendance 
		WHERE event_id = ? AND LOWER(attendee_name) = LOWER(?)
	`, event.ID, trimmedName).Scan(&existingID)

	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Duplicate Check-in: Attendee '%s' has already marked attendance for '%s'", trimmedName, event.Name),
		})
		return
	}

	// 2. Geofence distance calculation using Haversine formula
	distance := utils.HaversineDistance(req.Latitude, req.Longitude, event.Latitude, event.Longitude)

	// 3. Radius security check
	if distance > event.Radius {
		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("Attendance Rejected: You are %.1f meters away from the event location. Maximum permitted radius is %.1f meters.", distance, event.Radius),
			"distance_meters": distance,
			"allowed_radius":  event.Radius,
		})
		return
	}

	// 4. Save attendance
	now := time.Now()
	res, err := database.DB.Exec(`
		INSERT INTO attendance (event_id, attendee_name, latitude, longitude, distance, timestamp)
		VALUES (?, ?, ?, ?, ?, ?)
	`, event.ID, trimmedName, req.Latitude, req.Longitude, distance, now)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record attendance: " + err.Error()})
		return
	}

	attID, _ := res.LastInsertId()
	c.JSON(http.StatusOK, gin.H{
		"message":         "Attendance verified and recorded successfully!",
		"attendance_id":   attID,
		"event_name":      event.Name,
		"attendee_name":   trimmedName,
		"distance_meters": fmt.Sprintf("%.1f m", distance),
		"timestamp":       now.Format("2006-01-02 15:04:05"),
	})
}

func GetEventAttendance(c *gin.Context) {
	idParam := c.Param("id")
	eventID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	rows, err := database.DB.Query(`
		SELECT a.id, a.event_id, a.attendee_name, a.latitude, a.longitude, a.distance, a.timestamp, e.name
		FROM attendance a
		JOIN events e ON a.event_id = e.id
		WHERE a.event_id = ?
		ORDER BY a.timestamp DESC
	`, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query attendance"})
		return
	}
	defer rows.Close()

	list := make([]models.Attendance, 0)
	for rows.Next() {
		var a models.Attendance
		if err := rows.Scan(
			&a.ID, &a.EventID, &a.AttendeeName, &a.Latitude, &a.Longitude, &a.Distance, &a.Timestamp, &a.EventName,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan attendance row"})
			return
		}
		list = append(list, a)
	}

	c.JSON(http.StatusOK, list)
}

func ExportAttendanceCSV(c *gin.Context) {
	idParam := c.Param("id")
	eventID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var eventName string
	_ = database.DB.QueryRow("SELECT name FROM events WHERE id = ?", eventID).Scan(&eventName)
	if eventName == "" {
		eventName = "Event"
	}

	rows, err := database.DB.Query(`
		SELECT id, attendee_name, latitude, longitude, distance, timestamp
		FROM attendance
		WHERE event_id = ?
		ORDER BY timestamp DESC
	`, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query attendance"})
		return
	}
	defer rows.Close()

	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	_ = writer.Write([]string{
		"Attendance ID",
		"Event Name",
		"Attendee Name",
		"Latitude",
		"Longitude",
		"Distance from Center (m)",
		"Check-in Timestamp",
	})

	for rows.Next() {
		var id int64
		var name string
		var lat, lng, dist float64
		var ts time.Time

		if err := rows.Scan(&id, &name, &lat, &lng, &dist, &ts); err != nil {
			continue
		}

		_ = writer.Write([]string{
			strconv.FormatInt(id, 10),
			eventName,
			name,
			fmt.Sprintf("%.6f", lat),
			fmt.Sprintf("%.6f", lng),
			fmt.Sprintf("%.1f", dist),
			ts.Format("2006-01-02 15:04:05"),
		})
	}
	writer.Flush()

	filename := fmt.Sprintf("attendance_event_%d_%s.csv", eventID, time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "text/csv")
	c.Data(http.StatusOK, "text/csv", buf.Bytes())
}
