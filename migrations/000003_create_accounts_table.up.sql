CREATE TABLE accounts (
    id serial PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users ON DELETE CASCADE,
    balance numeric(20, 10) NOT NULL DEFAULT 0,
    withdrawn numeric(20, 10) NOT NULL DEFAULT 0
);

