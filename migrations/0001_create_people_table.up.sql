CREATE TABLE people (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    surname VARCHAR(100) NOT NULL,
    patronymic VARCHAR(100) DEFAULT NULL,
    age INTEGER CHECK (age >= 0 AND age <= 120),
    gender VARCHAR(50) CHECK (gender IN ('male', 'female', 'other')),
    nationality VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);