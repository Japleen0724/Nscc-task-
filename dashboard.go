package handlers

import (
	"net/http"

	"attendance-system/internal/database"
	"attendance-system/internal/models"
	"github.com/gin-gonic/gin"
)

func GetOrganizerDashboard(c *gin.Context) {
	organizerID := c.GetInt64("organizer_id")
	var metrics models.DashboardMetrics

	// Total events
	_ = database.DB.QueryRow(`
		SELECT COUNT(*) FROM events WHERE organizer_id = ?
	`, organizerID).Scan(&metrics.TotalEvents)

	// Total attendees
	_ = database.DB.QueryRow(`
		SELECT COUNT(a.id) 
		FROM attendance a
		JOIN events e ON a.event_id = e.id
		WHERE e.organizer_id = ?
	`, organizerID).Scan(&metrics.TotalAttendees)

	// Events summary
	rows, err := database.DB.Query(`
		SELECT e.id, e.name, COUNT(a.id) as attendee_count
		FROM events e
		LEFT JOIN attendance a ON e.id = a.event_id
		WHERE e.organizer_id = ?
		GROUP BY e.id
		ORDER BY e.created_at DESC
	`, organizerID)
	if err == nil {
		defer rows.Close()
		metrics.EventsSummary = make([]models.EventAttendanceSummary, 0)
		for rows.Next() {
			var s models.EventAttendanceSummary
			if err := rows.Scan(&s.EventID, &s.EventName, &s.AttendeeCount); err == nil {
				metrics.EventsSummary = append(metrics.EventsSummary, s)
			}
		}
	}

	// Recent check-ins
	checkinRows, err := database.DB.Query(`
		SELECT a.id, a.event_id, a.attendee_name, a.latitude, a.longitude, a.distance, a.timestamp, e.name
		FROM attendance a
		JOIN events e ON a.event_id = e.id
		WHERE e.organizer_id = ?
		ORDER BY a.timestamp DESC
		LIMIT 10
	`, organizerID)
	if err == nil {
		defer checkinRows.Close()
		metrics.RecentCheckins = make([]models.Attendance, 0)
		for checkinRows.Next() {
			var a models.Attendance
			if err := checkinRows.Scan(
				&a.ID, &a.EventID, &a.AttendeeName, &a.Latitude, &a.Longitude, &a.Distance, &a.Timestamp, &a.EventName,
			); err == nil {
				metrics.RecentCheckins = append(metrics.RecentCheckins, a)
			}
		}
	}

	c.JSON(http.StatusOK, metrics)
}
