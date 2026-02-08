-- name: MakeUserRed :one
update users 
set is_chirpy_red = true , updated_at = Now()
where id = $1
RETURNING *;