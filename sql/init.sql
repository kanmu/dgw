-- This file is automatically executed when the PostgreSQL container starts

-- The dgw_test user and database are created via environment variables
-- but we can grant additional privileges if needed
GRANT ALL PRIVILEGES ON DATABASE dgw_test TO dgw_test;

-- Create test tables for development
-- These are the same tables used in test.sql
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

-- Grant all privileges on tables to dgw_test user
GRANT ALL ON ALL TABLES IN SCHEMA public TO dgw_test;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO dgw_test;