-- name: GetChirpsByAuthor :many
select * from chirps 
where user_id = $1  
ORDER BY created_at;