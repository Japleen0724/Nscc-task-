package main

import (
	"log"
	"os"
	"path/filepath"

	"attendance-system/internal/database"
	"attendance-system/internal/handlers"
	"attendance-system/internal/middleware"
	"github.com/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Independent port from Library system (8080)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		if _, err := os.Stat("../database"); err == nil {
			dbPath = "../database/attendance.db"
		} else if _, err := os.Stat("database"); err == nil {
			dbPath = "database/attendance.db"
		} else {
			dbPath = filepath.Join("..", "database", "attendance.db")
		}
	}

	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	r := gin.Default()
	r.Use(corsMiddleware())

	// API Routes
	api := r.Group("/api")
	{
		// Auth
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.RegisterOrganizer)
			auth.POST("/login", handlers.LoginOrganizer)
			auth.GET("/me", middleware.RequireOrganizerAuth(), handlers.GetOrganizerMe)
		}

		// Public Event Lookup & QR
		api.GET("/events/lookup", handlers.LookupEventByQR)
		api.GET("/events/:id", handlers.GetEvent)
		api.GET("/events/:id/qr", handlers.GetEventQR)

		// Organizer Protected Routes
		events := api.Group("/events")
		events.Use(middleware.RequireOrganizerAuth())
		{
			events.GET("", handlers.ListEvents)
			events.POST("", handlers.CreateEvent)
			events.GET("/:id/attendance", handlers.GetEventAttendance)
			events.GET("/:id/export", handlers.ExportAttendanceCSV)
		}

		// Attendee Check-in (Public)
		api.POST("/attendance/checkin", handlers.CheckIn)

		// Organizer Dashboard
		organizer := api.Group("/organizer")
		organizer.Use(middleware.RequireOrganizerAuth())
		{
			organizer.GET("/dashboard", handlers.GetOrganizerDashboard)
		}
	}

	// Serve Frontend Static Files
	frontendDir := "../frontend"
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		frontendDir = "frontend"
	}
	if _, err := os.Stat(frontendDir); err == nil {
		r.Static("/frontend", frontendDir)
		r.StaticFile("/", filepath.Join(frontendDir, "index.html"))
		r.Static("/css", filepath.Join(frontendDir, "css"))
		r.Static("/js", filepath.Join(frontendDir, "js"))
	}

	log.Printf("==================================================")
	log.Printf("  QR Attendance System running on :%s", port)
	log.Printf("  Open in browser: http://localhost:%s", port)
	log.Printf("  Organizer: organizer@event.com / organizer123")
	log.Printf("==================================================")

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
