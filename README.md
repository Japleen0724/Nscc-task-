# QR-Based Geo-Tagged Attendance Management System

A full-stack, secure attendance management solution built with **Golang (Gin framework)**, **SQLite** (pure Go driver with zero CGO dependencies), and a modern **HTML5/CSS3/JavaScript** frontend with Geolocation and Geofencing.

Features real-time GPS verification using the spherical **Haversine formula**, duplicate check-in prevention, QR code generation/scanning, organizer dashboard with auto-refresh/polling, and CSV audit reports.

---

## Features

### 👨‍💼 Organizer Features
- **Authentication**: Organizer login (`organizer@event.com` / `organizer123`) and registration with JWT.
- **Event Creation**: Set event name, description, target GPS coordinates (latitude, longitude), and geofence radius (e.g. 100 meters).
- **GPS Auto-Detection**: One-click "Detect My Current GPS" button to automatically populate event coordinates from organizer's device.
- **QR Code Generation**: Dynamically generate high-resolution PNG QR codes for every event to display or download.
- **Real-Time Live Polling**: Automatic 5-second polling stream on the dashboard showing attendees marking check-ins live.
- **Attendance Audit**: View attendee records with exact coordinates, distance from center, and timestamps.
- **CSV Report Export**: Download full attendance audits as `.csv` files.

### 📱 Attendee Features
- **No App Installation Needed**: Works straight in any mobile or desktop web browser.
- **QR Code Verification**: Scan or input event QR code string (e.g. `EVENT-TECHCONF2026`).
- **Browser GPS Integration**: Securely requests device geolocation via HTML5 `navigator.geolocation.getCurrentPosition`.
- **Instant Geofence Feedback**: Live visual feedback showing calculated distance from event center in meters.
- **Duplicate Prevention**: Prevents duplicate attendance submissions by the same person for the same event.

### 🛡️ Attendance Security & Geofencing
- Calculates the spherical great-circle distance between attendee coordinates $(lat_1, lon_1)$ and event center $(lat_2, lon_2)$ using the **Haversine Formula**:
  $$d = 2 R \arcsin\left(\sqrt{\sin^2\left(\frac{\Delta \text{lat}}{2}\right) + \cos(\text{lat}_1)\cos(\text{lat}_2)\sin^2\left(\frac{\Delta \text{lon}}{2}\right)}\right)$$
  *(where Earth radius $R = 6,371,000$ meters)*.
- Strict server-side verification: Rejects any check-in attempt where $d > \text{radius}$ with `403 Forbidden`.
- Records actual detected distance and coordinates in SQLite database for full audit compliance.

---

## Project Structure

```
QR-Attendance-System/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Server entrypoint (runs on port 8081)
│   ├── internal/
│   │   ├── database/
│   │   │   └── db.go                # SQLite init, schema migrations, and auto-seeding
│   │   ├── handlers/
│   │   │   ├── attendance.go        # Check-in validation, Haversine verification, CSV export
│   │   │   ├── auth.go              # Organizer auth & JWT handlers
│   │   │   ├── dashboard.go         # Dashboard metrics & live check-ins feed
│   │   │   └── events.go            # Event creation, lookup, and QR generation
│   │   ├── middleware/
│   │   │   └── auth.go              # JWT authentication middleware
│   │   ├── models/
│   │   │   └── models.go            # Structs for Organizer, Event, Attendance, DTOs
│   │   └── utils/
│   │       └── geo.go               # Haversine distance algorithm implementation
│   ├── go.mod
│   └── go.sum
├── database/
│   ├── schema.sql                   # Raw SQLite schema definition
│   ├── seed.sql                     # Seed data with initial organizer, events, check-ins
│   └── attendance.db                # SQLite database (auto-created on launch)
├── frontend/
│   ├── css/
│   │   └── styles.css               # Modern responsive styling
│   ├── js/
│   │   ├── api.js                   # API client and auth handling
│   │   ├── app.js                   # Dashboard logic, event creation, polling
│   │   └── scanner.js               # Geolocation acquisition & attendee check-in
│   └── index.html                   # Dual-view interface (Attendee + Organizer)
└── README.md
```

---

## Prerequisites

- [Go](https://go.dev/dl/) version 1.20 or newer installed.
- Modern web browser (Chrome, Edge, Firefox, Safari) with location access enabled.

*Note: Uses `modernc.org/sqlite`, a 100% pure Go SQLite driver. No GCC, MinGW, or CGO toolchains are needed on Windows!*

---

## Getting Started

### 1. Navigate to the backend directory

```bash
cd QR-Attendance-System/backend
```

### 2. Initialize and download dependencies (already in go.mod)

```bash
go mod tidy
```

### 3. Run the application

```bash
go run ./cmd/server
```

Or build and execute the binary:
```bash
# Windows
go build -o attendance-server.exe ./cmd/server
.\attendance-server.exe

# Linux / macOS
go build -o attendance-server ./cmd/server
./attendance-server
```

### 4. Open the application

Open your browser and navigate to:
```
http://localhost:8081
```

*(Note: Uses port 8081 so it runs independently alongside the Library System on port 8080!)*

---

## Default Sample Credentials

| Role | Email | Password | Access |
| :--- | :--- | :--- | :--- |
| **Event Organizer** | `organizer@event.com` | `organizer123` | Create events, generate QR codes, live polling dashboard, export CSV |
| **Attendee** | None required | Public access | Scan event QR, grant location access, mark attendance |

*(You can also use the **Sign in as Demo Organizer** button in the modal to test immediately).*

---

## REST API Documentation

### Organizer Authentication
- `POST /api/auth/register` - Register a new event organizer.
- `POST /api/auth/login` - Authenticate and receive a JWT token.
- `GET /api/auth/me` - Get profile of authenticated organizer.

### Events
- `GET /api/events` *(Organizer)* - List all events created by organizer.
- `POST /api/events` *(Organizer)* - Create a new event with location coordinates and radius.
- `GET /api/events/:id` *(Public)* - Get public details of an event.
- `GET /api/events/:id/qr` *(Public)* - Download or render PNG QR code image.
- `GET /api/events/lookup?qr={qr_code}` *(Public)* - Lookup event by its QR code payload.

### Attendance & Verification
- `POST /api/attendance/checkin` *(Public)* - Submit attendance with attendee name, coordinates, and event ID/QR:
  - Validates duplicate check-ins.
  - Verifies attendee distance $\le \text{radius}$ using Haversine formula.
  - Returns calculated distance in meters.
- `GET /api/events/:id/attendance` *(Organizer)* - List attendee records for an event.
- `GET /api/events/:id/export` *(Organizer)* - Export attendance records as CSV.

### Dashboard
- `GET /api/organizer/dashboard` *(Organizer)* - Retrieve aggregate metrics and recent check-ins stream.
