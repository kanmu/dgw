DROP TABLE IF EXISTS t1;
DROP TABLE IF EXISTS t2;
DROP TABLE IF EXISTS t3;
DROP TABLE IF EXISTS t4;
DROP TABLE IF EXISTS t5;

CREATE TABLE t1 (
  id bigserial primary key
  , i integer not null unique
  , str text not null
  , nullable_str text
  , t_with_tz timestamp without time zone not null
  , t_without_tz timestamp with time zone not null
  , tm time
);

CREATE TABLE t2 (
  id bigserial primary key
  , i integer not null unique
  , str text not null
  , t_with_tz timestamp without time zone not null
  , t_without_tz timestamp with time zone not null
);

CREATE TABLE t3 (
  id bigserial not null
  , i integer not null
  , str text not null
  , t_with_tz timestamp without time zone not null
  , t_without_tz timestamp with time zone not null
  , PRIMARY KEY(id, i)
);

CREATE TABLE t4 (
  id integer not null
  , i integer not null
  , PRIMARY KEY(id, i)
);
