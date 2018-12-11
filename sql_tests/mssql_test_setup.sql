USE master
GO
IF NOT EXISTS (
   SELECT name
   FROM sys.databases
   WHERE name = N'sqlx'
)
CREATE DATABASE [sqlx]
GO