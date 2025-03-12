@echo off
setlocal

:: Set the service name
set SERVICE_NAME=easy-check

:: Set the executable path
set EXECUTABLE_PATH=%~dp0..\bin\easy-check.exe

:: Print debug information
echo Service name: %SERVICE_NAME%
echo Executable path: %EXECUTABLE_PATH%

:: Check if the executable file exists
if not exist "%EXECUTABLE_PATH%" (
  echo Executable file %EXECUTABLE_PATH% not found
  pause
  exit /b 1
)

:: Create a scheduled task to run the executable at startup
schtasks /create /tn "%SERVICE_NAME%" /tr "%EXECUTABLE_PATH%" /sc onlogon /rl highest

if %errorlevel% equ 0 (
  echo Scheduled task created and will run at startup
) else (
  echo Failed to create scheduled task
  pause
  exit /b 1
)

endlocal
pause
exit /b 0
