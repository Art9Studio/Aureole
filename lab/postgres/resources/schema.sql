CREATE TABLE orgs
(
    id   SERIAL PRIMARY KEY,
    name TEXT
);

CREATE TABLE users
(
    id       SERIAL PRIMARY KEY,
    username TEXT,
    phone    TEXT,
    email    TEXT,
    password TEXT,
    org_id   INT REFERENCES orgs
);

CREATE TABLE posts
(
    id      SERIAL PRIMARY KEY,
    content TEXT,
    user_id INT REFERENCES users
);

CREATE TABLE phone_verifications
(
    id       SERIAL PRIMARY KEY,
    phone    TEXT,
    code     TEXT,
    attempts INT,
    expires  timestamptz
);

CREATE TABLE password_resets
(
    id      SERIAL PRIMARY KEY,
    email   TEXT,
    token   TEXT,
    expires timestamptz,
    invalid BOOLEAN
);