CREATE TABLE users (
    id SERIAL  PRIMARY KEY,
    name VARCHAR DEFAULT 'noname_user',
    email VARCHAR UNIQUE NULL,
    pass TEXT NULL,
    telegram_id VARCHAR UNIQUE NULL,
    telegram_username VARCHAR UNIQUE NULL,
    max_id VARCHAR UNIQUE NULL,
    max_username VARCHAR UNIQUE NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)

CREATE INDEX idx_users_created_at ON users (created_at);

CREATE INDEX idx_users_name_email ON users (email);