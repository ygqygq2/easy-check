@echo off
setlocal

:: Set the service name
set SERVICE_NAME=easy-check

:: Delete the scheduled task
schtasks /delete /tn "%SERVICE_NAME%" /f

if %errorlevel% equ 0 (
  echo Scheduled task deleted
) else (
  echo Failed to delete scheduled task
  pause
  exit /b 1
)

endlocal
pause
exit /b 0
