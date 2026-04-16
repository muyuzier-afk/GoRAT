@echo off

set APP_NAME=factoryeye.exe
set APP_PATH=C:\ProgramData\Microsoft\Windows\%APP_NAME%

:: 检查ffmpeg是否安装
where ffmpeg >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未找到ffmpeg，请先安装ffmpeg并添加到系统PATH
    pause
    exit /b 1
)

:: 复制程序文件
copy /Y %APP_NAME% %APP_PATH%
if %errorlevel% neq 0 (
    echo 错误: 无法复制程序文件，请以管理员身份运行
    pause
    exit /b 1
)

:: 创建计划任务实现开机自启
schtasks /create /tn "FactoryEye" /tr "%APP_PATH%" /sc onstart /ru SYSTEM /f
if %errorlevel% neq 0 (
    echo 错误: 无法创建计划任务，请以管理员身份运行
    pause
    exit /b 1
)

:: 启动程序
start "" %APP_PATH%
echo 安装完成！程序已启动并设置为开机自启。
pause
