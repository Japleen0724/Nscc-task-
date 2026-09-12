-- Seed Data for QR-Based Geo-Tagged Attendance System
-- Organizer password is bcrypt hash for: organizer123

INSERT INTO organizers (id, name, email, password) VALUES
(1, 'Sarah Chen', 'organizer@event.com', '$2a$10$3Yd5O.3k8c/KjG5QyZp/Eev0mFf0eK.4oK2K8qR1q2T1zN3j1h2Gy');

INSERT INTO events (id, organizer_id, name, description, latitude, longitude, radius, qr_code, created_at) VALUES
(1, 1, 'Tech Conference 2026', 'Main Hall Keynote & Engineering Tracks', 37.7749, -122.4194, 150.0, 'EVENT-TECHCONF2026', datetime('now', '-5 days')),
(2, 1, 'Golang Developers Meetup', 'Hands-on concurrency and web architecture workshop', 12.9716, 77.5946, 100.0, 'EVENT-GOMEETUP26', datetime('now', '-2 days')),
(3, 1, 'Campus Career Fair', 'Auditorium building entrance registration', 40.7128, -74.0060, 200.0, 'EVENT-CAREERFAIR26', datetime('now', '-1 days'));

INSERT INTO attendance (id, event_id, attendee_name, latitude, longitude, distance, timestamp) VALUES
(1, 1, 'Michael Scott', 37.7750, -122.4193, 14.2, datetime('now', '-4 days')),
(2, 1, 'Pam Beesly', 37.7748, -122.4195, 12.8, datetime('now', '-4 days')),
(3, 1, 'Jim Halpert', 37.7751, -122.4192, 28.5, datetime('now', '-4 days')),
(4, 2, 'David Miller', 12.9717, 77.5947, 15.1, datetime('now', '-1 days')),
(5, 2, 'Priya Sharma', 12.9715, 77.5945, 16.3, datetime('now', '-1 days'));
