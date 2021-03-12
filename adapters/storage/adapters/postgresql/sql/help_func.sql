/* Returns the type of relationship between tables
   and the names of the fields that are used to establish relationships between tables
   IN:
        t1_name: first table name
        t2_name: second table name
   OUT:
        t1_names: first table field names
        t2_names: second table field names
        is_o2m: indicates whether relationship is one-to-many or not
*/
CREATE OR REPLACE FUNCTION get_rel_info(t1_name TEXT, t2_name TEXT)
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
