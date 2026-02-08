-- name: UpdateUser :one
update users 
set hashed_password = $2, email = $3, updated_at = Now()
where id = $1
RETURNING *;