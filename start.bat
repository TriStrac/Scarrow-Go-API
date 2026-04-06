@echo off
title Scarrow API - 1-Click Setup ^& Run
echo =======================================================
echo Preparing to run Scarrow API Setup Script...
echo =======================================================

:: Request Administrative privileges if not already granted (Optional, but good for Winget installations)
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo Requesting administrative privileges...
    powershell -Command "Start-Process '%~dpnx0' -Verb RunAs"
    exit /b
)

:: Bypass PowerShell execution policy and run the script
powershell.exe -ExecutionPolicy Bypass -NoProfile -File "%~dp0setup_and_run.ps1"

pause
