CREATE USER "go-crud-api" WITH PASSWORD 'go-crud-api';
CREATE DATABASE "go-crud-api";
GRANT ALL PRIVILEGES ON DATABASE "go-crud-api" to "go-crud-api";
\c "go-crud-api";
CREATE EXTENSION IF NOT EXISTS postgis;
