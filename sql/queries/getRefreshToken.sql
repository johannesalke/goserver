-- name: GetRefreshToken :one
Select * from refresh_tokens where token = $1;