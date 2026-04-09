@echo off
chcp 65001 >nul
setlocal

set "ROOT=%~dp0.."
pushd "%ROOT%"

set "GRAPH_SCRIPT=scripts\graph_tool.py"
set "DEFAULT_RESULT=data\output\result.csv"

if not exist "%GRAPH_SCRIPT%" (
    echo Error: %GRAPH_SCRIPT% not found.
    popd
    pause
    exit /b 1
)

if "%~1"=="" (
    echo No input files provided. Using default: %DEFAULT_RESULT%
    if not exist "%DEFAULT_RESULT%" (
        echo Error: %DEFAULT_RESULT% not found.
        popd
        pause
        exit /b 1
    )
    echo Running graph_tool.py for: %DEFAULT_RESULT%
    py "%GRAPH_SCRIPT%" "%DEFAULT_RESULT%"
) else (
    echo Using input files from command line:
    echo %*
    py "%GRAPH_SCRIPT%" %*
)

if errorlevel 1 (
    echo Error: graph_tool.py failed.
    popd
    pause
    exit /b 1
)

echo Done.
popd
pause
