@echo off
chcp 65001 >nul
setlocal

set "ROOT=%~dp0.."
pushd "%ROOT%"

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
echo Done.
echo Created files:
echo - %INPUT%
echo - %OUTPUT%
echo - %RESULT%
popd
pause
