CREATE TABLE users
(
    id       SERIAL,
    username TEXT,
    password TEXT,
    email    TEXT,
    org_id   INT REFERENCES orgs
);

ALTER TABLE users
    ADD CONSTRAINT "pk_key" PRIMARY KEY (id, username);

CREATE TABLE posts
(
    id       SERIAL PRIMARY KEY,
    content  TEXT,
    user_id  INT,
    username TEXT,
    FOREIGN KEY (user_id, username) REFERENCES users (id, username)
);

CREATE TABLE orgs
(
    id   SERIAL PRIMARY KEY,
    name TEXT
);

-- get all fk TO users
SELECT conrelid, conkey, confkey
FROM test.pg_catalog.pg_constraint
WHERE contype = 'f'
  AND confrelid = 'users'::regclass::oid;

-- get all fk FROM users
SELECT confrelid, conkey, confkey
FROM test.pg_catalog.pg_constraint
WHERE contype = 'f'
  AND conrelid = 'users'::regclass::oid;

SELECT relname
FROM pg_class
WHERE oid = (SELECT conrelid
             FROM test.pg_catalog.pg_constraint
             WHERE contype = 'f'
               AND confrelid = 'users'::regclass::oid);

SELECT attname
FROM test.pg_catalog.pg_attribute
WHERE attnum = 3
  AND attrelid = 'posts'::regclass::oid;

/* Возвращает тип связи между таблицами и имена полей, по которым устанавливаются связи между таблицами
   IN:
        t1_name: имя первой таблицы
        t2_name: имя второй таблиц
   OUT:
        t1_names: имена полей из первой таблицы
        t2_names: имена полей из второй таблицы
        is_o2m: один-ко-многим или нет
*/
CREATE OR REPLACE FUNCTION get_rel_fields_names(t1_name TEXT, t2_name TEXT)
    RETURNS TABLE
            (
                t1_names TEXT[],
                t2_names TEXT[],
                is_o2m   BOOLEAN
            )
AS
$$
DECLARE
constraint_rec RECORD;
    t1_oid         OID;
    t2_oid         OID;
    field_name     TEXT;
    field_key      SMALLINT;
    t1_field_keys  SMALLINT[];
    t2_field_keys  SMALLINT[];
BEGIN
SELECT *
INTO constraint_rec
FROM (
         SELECT conrelid, confrelid, conkey, confkey
         FROM test.pg_catalog.pg_constraint
         WHERE contype = 'f'
           AND confrelid = t1_name::regclass::oid
               AND conrelid = t2_name::regclass::oid

         UNION

         SELECT conrelid, confrelid, conkey, confkey
         FROM test.pg_catalog.pg_constraint
         WHERE contype = 'f'
           AND confrelid = t2_name::regclass::oid
           AND conrelid = t1_name::regclass::oid
     ) AS constr;

IF constraint_rec.confrelid = t1_name::regclass::oid THEN
        is_o2m = TRUE;
        t1_field_keys := t1_field_keys || constraint_rec.confkey;
        t2_field_keys := t2_field_keys || constraint_rec.conkey;
        t1_oid := constraint_rec.confrelid;
        t2_oid := constraint_rec.conrelid;
ELSE
        is_o2m = FALSE;
        t1_field_keys := t1_field_keys || constraint_rec.conkey;
        t2_field_keys := t2_field_keys || constraint_rec.confkey;
        t1_oid := constraint_rec.conrelid;
        t2_oid := constraint_rec.confrelid;
END IF;

    FOREACH field_key IN ARRAY t1_field_keys
        LOOP
            field_name := (SELECT attname
                           FROM test.pg_catalog.pg_attribute
                           WHERE attnum = field_key
                             AND attrelid = t1_oid);

            t1_names := array_append(t1_names, field_name);
END LOOP;

    FOREACH field_key IN ARRAY t2_field_keys
        LOOP
            field_name := (SELECT attname
                           FROM test.pg_catalog.pg_attribute
                           WHERE attnum = field_key
                             AND attrelid = t2_oid);

            t2_names := array_append(t2_names, field_name);
END LOOP;

    RETURN NEXT;
END;
$$
LANGUAGE plpgsql;

SELECT *
FROM get_rel_fields_names('users', 'posts');


