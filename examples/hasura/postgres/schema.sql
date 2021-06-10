CREATE TABLE orgs
(
    id   SERIAL PRIMARY KEY,
    name text
);

CREATE TABLE roles
(
    id SERIAL PRIMARY KEY,
    role TEXT
);

CREATE TABLE users
(
    id       SERIAL PRIMARY KEY,
    username TEXT,
    phone TEXT,
    email TEXT,
    password TEXT,
    org_id   INT REFERENCES orgs,
    role_id   INT REFERENCES roles
);

CREATE TABLE posts
(
    id       SERIAL PRIMARY KEY,
    content TEXT,
    user_id   INT REFERENCES users
);