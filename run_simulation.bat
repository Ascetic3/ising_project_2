@echo off
chcp 65001 >nul
setlocal

set PARAMS=params-sample2d.json
set INPUT=input.csv
set OUTPUT=output.csv
set RESULT=result.csv

echo [1/3] Generating input.csv...
if not exist "%PARAMS%" (
    echo Error: %PARAMS% not found.
    pause
    exit /b 1
)

if exist "%INPUT%" del "%INPUT%"
if exist "%OUTPUT%" del "%OUTPUT%"
if exist "%RESULT%" del "%RESULT%"

py task-parser3.py "%PARAMS%"
if not exist "%INPUT%" (
    echo Error: input.csv was not created.
    pause
    exit /b 1
)

echo [2/3] Running Go simulation...
go run ./cmd/run/main.go
if errorlevel 1 (
    echo Error: Go simulation failed.
    pause
    exit /b 1
)

if not exist "%OUTPUT%" (
    echo Error: output.csv was not created.
    pause
    exit /b 1
)

echo [3/3] Converting raw data...
py raw_data_converter.py "%OUTPUT%" "%RESULT%"
if errorlevel 1 (
    echo Error: raw_data_converter.py failed.
    pause
    exit /b 1
)

if not exist "%RESULT%" (
    echo Error: result.csv was not created.
    pause
    exit /b 1
)

echo Done.
echo Created files:
echo - %INPUT%
echo - %OUTPUT%
echo - %RESULT%
pause