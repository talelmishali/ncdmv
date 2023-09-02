-- name: GetAppointment :one
SELECT * FROM appointment
WHERE id = ? LIMIT 1;

-- name: GetAppointmentByLocationAndTime :one
SELECT * FROM appointment
WHERE location = ? AND time = ?
LIMIT 1;

-- name: ListAppointments :many
SELECT * FROM appointment
ORDER BY time DESC;

-- name: ListAppointmentsAfterDateForLocations :many
SELECT * FROM appointment
WHERE time >= ? AND location IN (sqlc.slice('locations'))
ORDER BY time DESC;

-- name: CreateAppointment :one
INSERT OR IGNORE INTO appointment (
  location, time, available
) VALUES (
  ?, ?, ?
)
RETURNING *;

-- name: UpdateAppointmentAvailable :exec
UPDATE appointment
SET available = ?
WHERE id = ?;

-- name: DeleteAppointment :exec
DELETE FROM appointment
WHERE id = ?;

-- name: PruneAppointmentsBeforeDate :many
UPDATE appointment
SET available = false
WHERE time < ? AND available = true
RETURNING *;

-- name: ListNotifications :many
SELECT * FROM notification;

-- name: CreateNotification :one
INSERT INTO notification (
  appointment_id, discord_webhook, available, appt_type
) VALUES (
  ?, ?, ?, ?
)
RETURNING *;

-- name: GetNotificationCountByAppointment :one
SELECT COUNT(*) FROM notification
WHERE appointment_id = ? AND discord_webhook = ?;
