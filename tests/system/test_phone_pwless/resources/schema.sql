CREATE TABLE users
(
    id             SERIAL PRIMARY KEY,
    username       TEXT,
    phone          TEXT,
    email          TEXT,
    password       TEXT,
    phone_verified BOOLEAN
);

CREATE TABLE phone_verifications
(
    id       SERIAL PRIMARY KEY,
    phone    TEXT,
    otp      TEXT,
    attempts INT,
    expires  timestamptz,
    invalid  BOOLEAN
);