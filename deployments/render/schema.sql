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

CREATE TABLE phone_verifications
(
    id       SERIAL PRIMARY KEY,
    phone    TEXT,
    otp      TEXT,
    attempts INT,
    expires  timestamptz,
    invalid  BOOLEAN
);

CREATE TABLE password_resets
(
    id      SERIAL PRIMARY KEY,
    email   TEXT,
    token   TEXT,
    expires timestamptz,
    invalid BOOLEAN
);

CREATE TABLE email_links
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

CREATE TABLE social_logins
(
    id        SERIAL PRIMARY KEY,
    social_id TEXT,
    email     TEXT,
    provider  TEXT,
    user_data jsonb,
    user_id   INT REFERENCES users
);