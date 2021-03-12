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


INSERT INTO orgs(name) VALUES ('Test'), ('Test 2');
INSERT INTO users(username, org_id) VALUES ('Test username', 1), ('Test username 2', 1), ('Test username 3', 2), ('Test username 4', 2);
INSERT INTO posts(user_id) VALUES (1), (1), (2), (2), (3), (3), (4), (4);