CREATE TABLE t1 (
  id bigserial primary key
  , i integer not null unique
  , str text not null
  , num_float numeric not null
  , nullable_str text
  , t_with_tz timestamp without time zone not null
  , t_without_tz timestamp with time zone not null
  , nullable_tz timestamp with time zone
  , json_data json not null
  , xml_data xml not null
  , tm time
);

CREATE TABLE t2 (
  id bigserial not null
  , i integer not null
  , str text not null
  , t_with_tz timestamp without time zone not null
  , t_without_tz timestamp with time zone not null
  , PRIMARY KEY(id, i)
);

CREATE TABLE t3 (
  id integer not null
  , i integer not null
  , PRIMARY KEY(id, i)
);
