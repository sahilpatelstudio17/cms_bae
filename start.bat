@echo off
setlocal enabledelayedexpansion

set DATABASE_URL=postgresql://postgres:new123@localhost:5432/cms_saas?sslmode=disable
set APP_PORT=8080

cd /d "%~dp0"
go run cmd\server\main.go
