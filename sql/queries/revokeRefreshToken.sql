-- name: RevokeRefreshToken :execresult
update refresh_tokens 
set revoked_at = NOW(), updated_at = NOW()
where token = $1;