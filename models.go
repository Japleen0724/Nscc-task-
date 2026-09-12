package models

import "time"

type Organizer struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Event struct {
	ID          int64     `json:"id"`
	OrganizerID int64     `json:"organizer_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Radius      float64   `json:"radius"` // Allowed perimeter radius in meters
	QRCode      string    `json:"qr_code"`
	CreatedAt   time.Time `json:"created_at"`

	// Joined stats
	AttendeeCount int64 `json:"attendee_count"`
}

type Attendance struct {
	ID           int64     `json:"id"`
	EventID      int64     `json:"event_id"`
	AttendeeName string    `json:"attendee_name"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Distance     float64   `json:"distance"` // Meters from event center
	Timestamp    time.Time `json:"timestamp"`

	// Joined fields
	EventName string `json:"event_name,omitempty"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=4"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	Organizer Organizer `json:"organizer"`
}

type CreateEventRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude" binding:"required"`
	Longitude   float64 `json:"longitude" binding:"required"`
	Radius      float64 `json:"radius"` // Optional, default 100 meters
}

type CheckinRequest struct {
	EventID      int64   `json:"event_id"`
	QRCode       string  `json:"qr_code"`
	AttendeeName string  `json:"attendee_name" binding:"required"`
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
}

type EventAttendanceSummary struct {
	EventID       int64  `json:"event_id"`
	EventName     string `json:"event_name"`
	AttendeeCount int64  `json:"attendee_count"`
}

type DashboardMetrics struct {
	TotalEvents     int64                    `json:"total_events"`
	TotalAttendees  int64                    `json:"total_attendees"`
	EventsSummary   []EventAttendanceSummary `json:"events_summary"`
	RecentCheckins  []Attendance             `json:"recent_checkins"`
}
