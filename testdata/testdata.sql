INSERT INTO users (login, PASSWORD)
    VALUES ('gopher', '$2a$14$exSjgqssYnKcJdJY0wJcTeqdpdrH7e4tz8wM3brPZaVtoDV/75UW6'),
    ('is', '$2a$14$eI49AGik17xObotoJU3C5OUpycqYVqlqH.vtEU6XiuoTfwmUcv28m'),
    ('very', '$2a$14$Xk/UB06CnXOfM4CSc1nqlesklfaK9QW7txIEsnekuzX8BJWTBkq9S'),
    ('happy', '$2a$14$yW/o.msyGbKB1YKj6nKn3OXOvoHbQhFnyc35e73WLzBobuWFPZOIm')
ON CONFLICT
    DO NOTHING;

INSERT INTO orders (user_id, number)
    VALUES (1, '79927398713'),
    (2, '49927398716'),
    (2, '0')
ON CONFLICT
    DO NOTHING;

