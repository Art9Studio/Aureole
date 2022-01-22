CREATE TABLE users
(
    id             serial PRIMARY KEY,
    username       VARCHAR UNIQUE,
    phone          VARCHAR UNIQUE,
    email          VARCHAR UNIQUE,
    email_verified boolean DEFAULT FALSE,
    phone_verified boolean DEFAULT FALSE,
    additional     jsonb
);

CREATE TABLE passwords
(
    id       serial PRIMARY KEY,
    user_id  INT REFERENCES users ON DELETE CASCADE,
    password VARCHAR

);

CREATE TABLE social_providers
(
    id            serial PRIMARY KEY,
    user_id       INT REFERENCES users ON DELETE CASCADE,
    provider_name VARCHAR,
    payload       jsonb
);

CREATE TABLE mfa
(
    id       serial PRIMARY KEY,
    user_id  INT REFERENCES users ON DELETE CASCADE,
    mfa_name VARCHAR,
    payload  jsonb
);

---- create above / drop below ----

DROP TABLE users;
DROP TABLE passwords;
DROP TABLE social_providers;
DROP TABLE mfa;
