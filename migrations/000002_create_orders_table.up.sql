CREATE TYPE order_status AS ENUM (
    'INVALID',
    'PROCESSED',
    'NEW',
    'PROCESSING'
);

CREATE TABLE orders (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users ON DELETE CASCADE,
    number integer NOT NULL CONSTRAINT unique_order_number UNIQUE,
    status order_status NOT NULL DEFAULT 'NEW',
    uploadet_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

