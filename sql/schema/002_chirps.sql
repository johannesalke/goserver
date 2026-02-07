-- +goose Up
CREATE TABLE chirps (
    id UUID primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    body text not null,
    user_id UUID not null REFERENCES users ON DELETE CASCADE


);

-- +goose Down
DROP TABLE chirps;