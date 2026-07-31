@echo off
setlocal EnableExtensions

set "ROOT=%~dp0.."
pushd "%ROOT%"
if errorlevel 1 exit /b 1

set "DEMO_SEED=20260731"
for /f %%I in ('powershell -NoProfile -Command "Get-Date -Format yyyyMMdd_HHmmss_fff"') do set "RUN_ID=%%I"
if not defined RUN_ID set "RUN_ID=%RANDOM%%RANDOM%"
set "OUTPUT_DIR=demo-output\metropolis_%RUN_ID%_seed_%DEMO_SEED%"
set "RUNNER=%TEMP%\ising_metropolis_demo_%RANDOM%%RANDOM%.exe"

go build -o "%RUNNER%" ./cmd/run
if errorlevel 1 goto :fail

"%RUNNER%" -input configs\demo-input.csv -output-dir "%OUTPUT_DIR%" -seed %DEMO_SEED% -save-images
if errorlevel 1 goto :fail

if exist "%RUNNER%" del "%RUNNER%"
echo Demo results: %CD%\%OUTPUT_DIR%
popd
exit /b 0

:fail
set "EXIT_CODE=%ERRORLEVEL%"
if exist "%RUNNER%" del "%RUNNER%"
echo Demo failed with exit code %EXIT_CODE%.
popd
exit /b %EXIT_CODE%
