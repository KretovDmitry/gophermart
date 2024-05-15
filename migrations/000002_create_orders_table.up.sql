CREATE TYPE order_status AS ENUM (
    'INVALID',
    'PROCESSED',
    'NEW',
    'PROCESSING'
);

CREATE TABLE orders (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users ON DELETE RESTRICT,
    number text NOT NULL CONSTRAINT unique_order_number UNIQUE,
    status order_status NOT NULL DEFAULT 'NEW',
    accrual numeric(20, 10) NOT NULL DEFAULT 0,
    uploadet_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

