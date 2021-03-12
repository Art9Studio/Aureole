select username from users where id=1; -- field, no keys
select row_to_json(t) from (select username from users where id=1) t; -- field, keys

select json_agg(t.id) from (select p.id from users join posts p on users.id = p.user_id where users.id=1) t; -- field_array, no keys
select json_agg(t) from (select p.id from users join posts p on users.id = p.user_id where users.id=1) t; -- field_array, keys

select json_build_array(username, password) from users where id=1; -- row, no keys
select row_to_json(t) from (select username, password from users where id=1) t; -- (row, keys) === (field, keys)

select json_build_array(p.id, p.content)  from users join posts p on users.id = p.user_id where users.id=1; -- row_array, no keys
select json_agg(t) from (select p.id, p.content from users join posts p on users.id = p.user_id where users.id=1) t; -- row_array, keys