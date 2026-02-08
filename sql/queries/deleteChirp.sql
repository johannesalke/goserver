-- name: DeleteChirp :execresult
DELETE FROM chirps
where id = $1;