INSERT INTO users (login, PASSWORD)
    VALUES ('gopher', '$2a$14$exSjgqssYnKcJdJY0wJcTeqdpdrH7e4tz8wM3brPZaVtoDV/75UW6'),
    ('is', '$2a$14$eI49AGik17xObotoJU3C5OUpycqYVqlqH.vtEU6XiuoTfwmUcv28m'),
    ('very', '$2a$14$Xk/UB06CnXOfM4CSc1nqlesklfaK9QW7txIEsnekuzX8BJWTBkq9S'),
    ('happy', '$2a$14$yW/o.msyGbKB1YKj6nKn3OXOvoHbQhFnyc35e73WLzBobuWFPZOIm')
ON CONFLICT
    DO NOTHING;

INSERT INTO orders (user_id, number, status, accrual)
    VALUES (1, '79927398713', 'NEW', 0),
    (2, '49927398716', 'PROCESSED', 417.863),
    (2, '0', 'PROCESSING', 0),
    (1, '7992739871', 'NEW', 0),
    (1, '7992739873', 'NEW', 0),
    (1, '7992739813', 'NEW', 0),
    (1, '7992739713', 'NEW', 0),
    (1, '7992738713', 'NEW', 0),
    (1, '7992798713', 'NEW', 0),
    (1, '7992398713', 'NEW', 0),
    (1, '7997398713', 'NEW', 0),
    (1, '7927398713', 'NEW', 0),
    (1, '7927398713', 'NEW', 0),
    (1, '9927398713', 'NEW', 0)
ON CONFLICT
    DO NOTHING;

INSERT INTO accounts (user_id, balance, withdrawn)
    VALUES (1, 589.109, 703.12),
    (2, 113.14, 901.07),
    (3, 1064, 0)
ON CONFLICT
    DO NOTHING;

INSERT INTO account_operations (account_id, operation, order_number, sum)
    VALUES (1, 'WITHDRAWAL', '79927398713', '17.309'),
    (2, 'ACCRUAL', '49927398716', '47.001'),
    (2, 'WITHDRAWAL', '0', '980.13552')
ON CONFLICT
    DO NOTHING;

