-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS shorten_url (
    id VARCHAR PRIMARY KEY,
    url VARCHAR,
    deleted_flag boolean DEFAULT false
);

CREATE UNIQUE INDEX IF NOT EXISTS url_idx ON shorten_url (url);

CREATE TABLE IF NOT EXISTS users (
	id VARCHAR PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users_shorten_url (
    user_id VARCHAR,
    url_id VARCHAR,

    PRIMARY KEY(user_id, url_id),

    CONSTRAINT fk_user_id
    FOREIGN KEY (user_id) 
    REFERENCES users (id),

    CONSTRAINT fk_url_id
    FOREIGN KEY (url_id) 
    REFERENCES shorten_url (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users_shorten_url;
DROP TABLE IF EXISTS users;
DROP INDEX IF EXISTS url_idx;
DROP TABLE IF EXISTS shorten_url;
-- +goose StatementEnd
