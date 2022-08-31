package migrations

const upSchema1 = `
CREATE TABLE users
(
    id             SERIAL PRIMARY KEY,
    username       VARCHAR UNIQUE,
    phone          VARCHAR UNIQUE,
    email          VARCHAR UNIQUE,
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
	is_mfa_enabled BOOLEAN,
	enabled_mfas   VARCHAR[]
);

CREATE TABLE imported_users
(
	user_id INT   REFERENCES users ON DELETE CASCADE,
	plugin_id     VARCHAR NOT NULL,
	provider_name VARCHAR NOT NULL,
	provider_id   VARCHAR NOT NULL,
	additional    jsonb NOT NULL,
	UNIQUE (user_id, plugin_id)
);

CREATE TABLE secrets
(
	user_id INT REFERENCES users ON DELETE CASCADE,
	plugin_id VARCHAR NOT NULL,
	payload jsonb NOT NULL,
	UNIQUE (user_id, plugin_id)
);

CREATE TABLE plugins
(
	id VARCHAR UNIQUE NOT NULL,
	plugin_name VARCHAR NOT NULL
);
`

const downSchema1 = `
DROP TABLE imported_users;
DROP TABLE secrets;
DROP TABLE users;
DROP TABLE plugins;
`

func init() {
	appendMigration(upSchema1, downSchema1)
}
