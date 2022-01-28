CREATE DATABASE [go-crud-api]
GO
CREATE LOGIN [go-crud-api] WITH PASSWORD=N'go-crud-api', DEFAULT_DATABASE=[go-crud-api], CHECK_EXPIRATION=OFF, CHECK_POLICY=OFF
GO
USE [go-crud-api]
GO
CREATE USER [go-crud-api] FOR LOGIN [go-crud-api] WITH DEFAULT_SCHEMA=[dbo]
exec sp_addrolemember 'db_owner', 'go-crud-api';
GO
