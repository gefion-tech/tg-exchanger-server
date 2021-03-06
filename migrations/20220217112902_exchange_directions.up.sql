CREATE TABLE IF NOT EXISTS exchange_directions(
    id BIGSERIAL NOT NULL PRIMARY KEY,
    exchange_from VARCHAR(20) NOT NULL,
    exchange_to VARCHAR(20) NOT NULL,
    course_correction INT NOT NULL,
    address_verification BOOLEAN DEFAULT false,
    status BOOLEAN DEFAULT false,
    created_by VARCHAR(20) REFERENCES users(username),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),

    UNIQUE (exchange_from, exchange_to)
);