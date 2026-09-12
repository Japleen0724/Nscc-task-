package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	log.Printf("Connected to SQLite database at: %s", dbPath)

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	if err := seedDefaults(db); err != nil {
		return nil, fmt.Errorf("failed to seed defaults: %w", err)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS organizers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		organizer_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		radius REAL NOT NULL DEFAULT 100.0,
		qr_code TEXT UNIQUE NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (organizer_id) REFERENCES organizers(id)
	);

	CREATE TABLE IF NOT EXISTS attendance (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id INTEGER NOT NULL,
		attendee_name TEXT NOT NULL,
		latitude REAL NOT NULL,
		longitude REAL NOT NULL,
		distance REAL NOT NULL DEFAULT 0.0,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (event_id) REFERENCES events(id),
		UNIQUE (event_id, attendee_name)
	);

	CREATE INDEX IF NOT EXISTS idx_events_qr ON events(qr_code);
	CREATE INDEX IF NOT EXISTS idx_events_organizer ON events(organizer_id);
	CREATE INDEX IF NOT EXISTS idx_attendance_event ON attendance(event_id);
	CREATE INDEX IF NOT EXISTS idx_attendance_timestamp ON attendance(timestamp);
	`

	_, err := db.Exec(schema)
	return err
}

func seedDefaults(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM organizers").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("organizer123"), bcrypt.DefaultCost)
		_, err = db.Exec(`
			INSERT INTO organizers (id, name, email, password) VALUES 
			(1, 'Sarah Chen', 'organizer@event.com', ?)
		`, string(hashed))
		if err != nil {
			return err
		}
		log.Println("Seeded default organizer (organizer@event.com / organizer123)")
	}

	var eventCount int
	err = db.QueryRow("SELECT COUNT(*) FROM events").Scan(&eventCount)
	if err != nil {
		return err
	}

	if eventCount == 0 {
		_, err = db.Exec(`
			INSERT INTO events (id, organizer_id, name, description, latitude, longitude, radius, qr_code, created_at) VALUES 
			(1, 1, 'Tech Conference 2026', 'Main Hall Keynote & Engineering Tracks', 37.7749, -122.4194, 150.0, 'EVENT-TECHCONF2026', datetime('now', '-5 days')),
			(2, 1, 'Golang Developers Meetup', 'Hands-on concurrency and web architecture workshop', 12.9716, 77.5946, 100.0, 'EVENT-GOMEETUP26', datetime('now', '-2 days')),
			(3, 1, 'Campus Career Fair', 'Auditorium building entrance registration', 40.7128, -74.0060, 200.0, 'EVENT-CAREERFAIR26', datetime('now', '-1 days'));
		`)
		if err != nil {
			return err
		}
		log.Println("Seeded default events")
	}

	var attendanceCount int
	err = db.QueryRow("SELECT COUNT(*) FROM attendance").Scan(&attendanceCount)
	if err != nil {
		return err
	}

	if attendanceCount == 0 {
		_, err = db.Exec(`
			INSERT INTO attendance (id, event_id, attendee_name, latitude, longitude, distance, timestamp) VALUES 
			(1, 1, 'Michael Scott', 37.7750, -122.4193, 14.2, datetime('now', '-4 days')),
			(2, 1, 'Pam Beesly', 37.7748, -122.4195, 12.8, datetime('now', '-4 days')),
			(3, 1, 'Jim Halpert', 37.7751, -122.4192, 28.5, datetime('now', '-4 days')),
			(4, 2, 'David Miller', 12.9717, 77.5947, 15.1, datetime('now', '-1 days')),
			(5, 2, 'Priya Sharma', 12.9715, 77.5945, 16.3, datetime('now', '-1 days'));
		`)
		if err != nil {
			return err
		}
		log.Println("Seeded default attendance records")
	}

	return nil
}
