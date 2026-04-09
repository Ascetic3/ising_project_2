@echo off
chcp 65001 >nul
setlocal EnableExtensions

set "ROOT=%~dp0.."
pushd "%ROOT%"
set "START_TIME=%TIME%"

set "PARAMS=configs\params-sample2d.json"
set "INPUT=data\input\input.csv"
set "OUTPUT=data\output\output.csv"
set "RESULT=data\output\result.csv"
set "GRAPH_SCRIPT=scripts\graph_tool.py"
set "PLOTS_DIR=data\output\plots"

if not exist "data\input" mkdir "data\input"
if not exist "data\output" mkdir "data\output"
if not exist "%PLOTS_DIR%" mkdir "%PLOTS_DIR%"

echo [1/4] Generating input.csv...
if not exist "%PARAMS%" (
    echo Error: %PARAMS% not found.
    popd
    pause
    exit /b 1
)

if exist "%INPUT%" del "%INPUT%"
if exist "%OUTPUT%" del "%OUTPUT%"
if exist "%RESULT%" del "%RESULT%"

py scripts\make_input_csv.py "%PARAMS%"
if not exist "%INPUT%" (
    echo Error: input.csv was not created.
    popd
    pause
    exit /b 1
)

echo [2/4] Running Go simulation...
go run ./cmd/run/main.go
if errorlevel 1 (
    echo Error: Go simulation failed.
    popd
    pause
    exit /b 1
)

if not exist "%OUTPUT%" (
    echo Error: output.csv was not created.
    popd
    pause
    exit /b 1
)

echo [3/4] Converting raw data...
py scripts\make_result_csv.py "%OUTPUT%" "%RESULT%"
if errorlevel 1 (
    echo Error: make_result_csv.py failed.
    popd
    pause
    exit /b 1
)

if not exist "%RESULT%" (
    echo Error: result.csv was not created.
    popd
    pause
    exit /b 1
)

echo [4/4] Graphs...
if not exist "%GRAPH_SCRIPT%" (
    echo Error: %GRAPH_SCRIPT% not found.
    popd
    pause
    exit /b 1
)
if not exist "%RESULT%" (
    echo Error: %RESULT% not found.
    popd
    pause
    exit /b 1
)
echo Graph script: %GRAPH_SCRIPT%
echo Graph input : %RESULT%
echo Graph output: %PLOTS_DIR%

py -c "import matplotlib, numpy" >nul 2>&1
if errorlevel 1 (
    echo matplotlib/numpy not found. Installing...
    py -m pip install --user matplotlib numpy
    py -c "import matplotlib, numpy" >nul 2>&1
    if errorlevel 1 (
        echo Warning: could not install matplotlib/numpy. Skip graphs.
        goto :after_graphs
    )
)

echo Launching graphs in separate window...
start "Ising Graphs" cmd /k "cd /d %PLOTS_DIR% && py ..\..\..\%GRAPH_SCRIPT% ..\result.csv"
echo Graph process started.

:after_graphs
set "END_TIME=%TIME%"
for /f "tokens=1-4 delims=:.," %%a in ("%START_TIME%") do (
    set /a "START_CS=(((1%%a-100)*3600)+((1%%b-100)*60)+(1%%c-100))*100+(1%%d-100)"
)
for /f "tokens=1-4 delims=:.," %%a in ("%END_TIME%") do (
    set /a "END_CS=(((1%%a-100)*3600)+((1%%b-100)*60)+(1%%c-100))*100+(1%%d-100)"
)
if %END_CS% lss %START_CS% set /a "END_CS+=24*3600*100"
set /a "ELAPSED_CS=END_CS-START_CS"
set /a "ELAPSED_H=ELAPSED_CS/(3600*100)"
set /a "ELAPSED_M=(ELAPSED_CS/(60*100))%%60"
set /a "ELAPSED_S=(ELAPSED_CS/100)%%60"
set /a "ELAPSED_C=ELAPSED_CS%%100"
echo Done.
echo Created files:
echo - %INPUT%
echo - %OUTPUT%
echo - %RESULT%
echo Start time: %START_TIME%
echo End time  : %END_TIME%
echo Elapsed   : %ELAPSED_H%h %ELAPSED_M%m %ELAPSED_S%s %ELAPSED_C%cs
popd
pause
