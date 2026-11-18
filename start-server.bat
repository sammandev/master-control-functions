@echo off
set DATABASE_URL=postgresql://postgres:pega1234@127.0.0.1:5432/master_control_functions?sslmode=disable
cd /d "%~dp0"
echo Starting Master Control Functions on http://localhost:8088 ...
backend\webprogfndb-server.exe -addr 127.0.0.1:8088 -web frontend/build
pause
