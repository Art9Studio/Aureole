CREATE TABLE orgs
(
    id   SERIAL PRIMARY KEY,
    name text
);

CREATE TABLE users
(
    id       SERIAL PRIMARY KEY,
    username TEXT,
    password text,
    org_id   int REFERENCES orgs
);

CREATE TABLE posts
(
    id       SERIAL PRIMARY KEY,
    content TEXT,
    user_id   int REFERENCES users
);