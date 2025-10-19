CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    balance NUMERIC(10,2) DEFAULT 0
);

INSERT INTO users (name, email, balance) VALUES
('Adilet', 'adilet@example.com', 1000),
('Ali', 'ali@example.com', 500),
('Dan', 'dan@example.com', 200);
