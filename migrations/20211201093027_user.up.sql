CREATE TABLE users (
    chat_id BIGINT NOT NULL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    hash VARCHAR(255),
    role INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);