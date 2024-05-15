CREATE TYPE account_operation AS enum (
    'ACCRUAL',
    'WITHDRAWAL'
);

CREATE TABLE account_operations (
    id bigserial PRIMARY KEY,
    account_id integer NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    operation account_operation NOT NULL,
    order_number text NOT NULL REFERENCES orders (number) ON DELETE RESTRICT,
    sum numeric(20, 10) NOT NULL,
    processed_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

