CREATE DATABASE `go-crud-api` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE USER 'go-crud-api'@'%'  IDENTIFIED BY 'go-crud-api';
GRANT ALL PRIVILEGES ON `go-crud-api`.* TO 'go-crud-api'@'%' WITH GRANT OPTION;
FLUSH PRIVILEGES;
