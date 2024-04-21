CREATE TABLE users (
    id serial PRIMARY KEY,
    login varchar(255) NOT NULL CONSTRAINT unique_login UNIQUE,
    password varchar(255) NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

