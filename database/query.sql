-- name: GetAppointment :one
SELECT * FROM appointment
WHERE id = ? LIMIT 1;

-- name: ListAppointments :many
SELECT * FROM appointment
ORDER BY time DESC;

-- name: CreateAppointment :one
INSERT INTO appointment (
  location, time
) VALUES (
  ?, ?
)
RETURNING *;

-- name: DeleteAppointment :exec
DELETE FROM appointment
WHERE id = ?;
