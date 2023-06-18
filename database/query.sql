-- name: GetAppointment :one
SELECT * FROM appointment
WHERE id = ? LIMIT 1;

-- name: ListAppointments :many
SELECT * FROM appointment
ORDER BY time DESC;

-- name: CreateAppointment :one
INSERT OR IGNORE INTO appointment (
  location, time
) VALUES (
  ?, ?
)
RETURNING *;

-- name: DeleteAppointment :exec
DELETE FROM appointment
WHERE id = ?;

-- name: ListNotifications :many
SELECT * FROM notification;

-- name: CreateNotification :one
INSERT INTO notification (
  appointment_id, discord_webhook
) VALUES (
  ?, ?
)
RETURNING *;

-- name: GetNotificationCountByAppointment :one
SELECT COUNT(*) FROM notification
WHERE appointment_id = ? AND discord_webhook = ?;
