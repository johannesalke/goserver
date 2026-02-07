-- name: GetChirps :many
select * from chirps ORDER BY created_at;