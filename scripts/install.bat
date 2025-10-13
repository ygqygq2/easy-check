@echo off@echo off

chcp 65001 >nulchcp 65001 >nul

setlocal enabledelayedexpansionsetlocal



echo ========================================echo ========================================

echo Easy-Check 开机自启安装echo Easy-Check 开机自启安装工具（计划任务方式）

echo ========================================echo ========================================

echo.echo.



:select_version:: Set the service name

echo 请选择要安装的版本:set SERVICE_NAME=easy-check

echo 1. UI 版本 (easy-check-ui-windows-amd64.exe)

echo 2. CMD 版本 (easy-check-windows-amd64.exe):: Set the executable path

echo 3. 退出set EXECUTABLE_PATH=%~dp0..\bin\easy-check-ui-windows-amd64.exe

echo.

set /p choice=请输入选项 (1-3): :: Print debug information

echo 服务名称: %SERVICE_NAME%

if "%choice%"=="1" goto install_uiecho 可执行文件: %EXECUTABLE_PATH%

if "%choice%"=="2" goto install_cmdecho.

if "%choice%"=="3" goto end

echo [错误] 无效的选项，请重新选择:: Check if the executable file exists

echo.if not exist "%EXECUTABLE_PATH%" (

goto select_version  echo [错误] 找不到可执行文件: %EXECUTABLE_PATH%

  echo 请确保程序已正确编译并放置在 bin 目录下

:install_ui  pause

set APP_NAME=Easy-Check-UI  exit /b 1

set EXECUTABLE_PATH=%~dp0..\bin\easy-check-ui-windows-amd64.exe)

goto install

:: 检查是否已存在同名任务，如果存在则先删除

:install_cmdschtasks /query /tn "%SERVICE_NAME%" >nul 2>&1

set APP_NAME=Easy-Check-CMDif %errorlevel% equ 0 (

set EXECUTABLE_PATH=%~dp0..\bin\easy-check-windows-amd64.exe  echo [提示] 检测到已存在的计划任务，将先删除...

goto install  schtasks /delete /tn "%SERVICE_NAME%" /f >nul 2>&1

)

:install

set STARTUP_FOLDER=%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup:: Create a scheduled task to run the executable at startup

set SHORTCUT_NAME=%APP_NAME%.lnk:: /rl highest: 以最高权限运行

set SHORTCUT_PATH=%STARTUP_FOLDER%\%SHORTCUT_NAME%:: /sc onlogon: 用户登录时运行

:: /f: 强制创建（覆盖已存在的）

echo.echo 正在创建计划任务...

echo 程序名称: %APP_NAME%schtasks /create /tn "%SERVICE_NAME%" /tr "\"%EXECUTABLE_PATH%\"" /sc onlogon /rl highest /f

echo 可执行文件: %EXECUTABLE_PATH%

echo 快捷方式路径: %SHORTCUT_PATH%if %errorlevel% equ 0 (

echo.  echo.

  echo [成功] 计划任务创建成功，程序将在用户登录时自动启动

if not exist "%EXECUTABLE_PATH%" (  echo.

    echo [错误] 找不到可执行文件: %EXECUTABLE_PATH%  echo 任务名称: %SERVICE_NAME%

    echo 请确保程序已正确编译  echo 运行权限: 最高权限

    pause  echo 触发条件: 用户登录时

    exit /b 1  echo.

)  echo 如需取消开机自启，请运行 uninstall.bat

) else (

if exist "%SHORTCUT_PATH%" (  echo.

    echo [提示] 快捷方式已存在，将被覆盖  echo [错误] 创建计划任务失败

    del "%SHORTCUT_PATH%"  echo 请确保以管理员权限运行此脚本

)  pause

  exit /b 1

echo 正在创建快捷方式...)

powershell -NoProfile -Command "$WS = New-Object -ComObject WScript.Shell; $SC = $WS.CreateShortcut('%SHORTCUT_PATH%'); $SC.TargetPath = '%EXECUTABLE_PATH%'; $SC.WorkingDirectory = '%~dp0..'; $SC.Description = '%APP_NAME% - 网络检测工具'; $SC.Save()"

endlocal

if %errorlevel% equ 0 (echo.

    echo.pause

    echo [成功] %APP_NAME% 开机自启已设置成功！exit /b 0

    echo 程序将在下次登录时自动启动
) else (
    echo.
    echo [错误] 创建快捷方式失败
    pause
    exit /b 1
)

echo.
pause
exit /b 0

:end
echo.
echo 已取消安装
exit /b 0
