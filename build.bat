@echo off

cd %~dp0
go build

if not "%ERRORLEVEL%"=="0" (
	goto :eof
)

setlocal
set PORT=3001
heatmap.exe
