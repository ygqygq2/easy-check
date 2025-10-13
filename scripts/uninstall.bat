@echo off@echo off

chcp 65001 >nulchcp 65001 >nul

setlocal enabledelayedexpansionsetlocal



echo ========================================echo ========================================

echo Easy-Check 开机自启卸载echo Easy-Check 开机自启卸载工具（计划任务方式）

echo ========================================echo ========================================

echo.echo.



set STARTUP_FOLDER=%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup:: Set the service name

set SERVICE_NAME=easy-check

set UI_SHORTCUT=%STARTUP_FOLDER%\Easy-Check-UI.lnk

set CMD_SHORTCUT=%STARTUP_FOLDER%\Easy-Check-CMD.lnkecho 服务名称: %SERVICE_NAME%

echo.

set FOUND=0

:: 检查计划任务是否存在

if exist "%UI_SHORTCUT%" (schtasks /query /tn "%SERVICE_NAME%" >nul 2>&1

    echo 正在删除 UI 版本快捷方式...if %errorlevel% neq 0 (

    del "%UI_SHORTCUT%"  echo [提示] 计划任务不存在，可能已经被删除

    if %errorlevel% equ 0 (  echo.

        echo [成功] UI 版本快捷方式已删除  pause

        set FOUND=1  exit /b 0

    ))

)

:: Delete the scheduled task

if exist "%CMD_SHORTCUT%" (echo 正在删除计划任务...

    echo 正在删除 CMD 版本快捷方式...schtasks /delete /tn "%SERVICE_NAME%" /f

    del "%CMD_SHORTCUT%"

    if %errorlevel% equ 0 (if %errorlevel% equ 0 (

        echo [成功] CMD 版本快捷方式已删除  echo.

        set FOUND=1  echo [成功] 计划任务已删除

    )  echo 程序不会在下次登录时自动启动

)) else (

  echo.

echo.  echo [错误] 删除计划任务失败

if %FOUND%==1 (  echo 请确保以管理员权限运行此脚本

    echo [完成] 开机自启已取消  pause

) else (  exit /b 1

    echo [提示] 未找到任何开机自启项)

)

endlocal

echo.echo.

pausepause

exit /b 0exit /b 0
