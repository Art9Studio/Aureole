CREATE TABLE users
(
    id             SERIAL PRIMARY KEY,
    username       TEXT,
    phone          TEXT,
    email          TEXT,
    password       TEXT,
    email_verified BOOLEAN
);

CREATE TABLE email_links
(
    id      SERIAL PRIMARY KEY,
    email   TEXT,
    token   TEXT,
    expires timestamptz,
    invalid BOOLEAN
);

CREATE TABLE sessions
(
    user_id    INT PRIMARY KEY REFERENCES users,
    session_id TEXT,
    expiration timestamptz
);