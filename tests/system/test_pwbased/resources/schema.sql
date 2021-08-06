CREATE TABLE users
(
    id       SERIAL PRIMARY KEY,
    username TEXT,
    phone    TEXT,
    email    TEXT,
    password TEXT,
    email_verified BOOLEAN
);

CREATE TABLE password_resets
(
    id      SERIAL PRIMARY KEY,
    email   TEXT,
    token   TEXT,
    expires timestamptz,
    invalid BOOLEAN
);

CREATE TABLE email_verifications
(
    id      SERIAL PRIMARY KEY,
    email   TEXT,
    token   TEXT,
    expires timestamptz,
    invalid BOOLEAN
);