CREATE TABLE users
(
    id             SERIAL PRIMARY KEY,
    username       TEXT,
    phone          TEXT,
    email          TEXT,
    password       TEXT,
    email_verified BOOLEAN,
    phone_verified BOOLEAN
);

CREATE TABLE social_logins
(
    id        SERIAL PRIMARY KEY,
    social_id TEXT,
    email     TEXT,
    provider  TEXT,
    user_data jsonb,
    user_id   INT REFERENCES users
);