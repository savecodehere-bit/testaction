@echo off
chcp 65001 >nul

REM 创建bin目录
if not exist "bin" mkdir bin

REM 检查是否已编译
if not exist "bin\center_service.exe" (
    echo 检测到可执行文件不存在，开始编译...
    call build.bat
)

echo.
echo 启动微服务（按顺序启动：注册中心 -^> 用户服务 -^> 订单服务 -^> 网关服务）...
echo 注意：每个服务会弹出独立的GUI窗口
echo.

REM 启动服务注册中心
echo 1. 启动服务注册中心（端口8080）...
start "服务注册中心" bin\center_service.exe
timeout /t 2 /nobreak >nul

REM 启动用户服务
echo 2. 启动用户服务（端口8081）...
start "用户服务" bin\user_service.exe
timeout /t 2 /nobreak >nul

REM 启动订单服务
echo 3. 启动订单服务（端口8082）...
start "订单服务" bin\order_service.exe
timeout /t 2 /nobreak >nul

REM 启动网关服务
echo 4. 启动API网关服务（端口8083）...
start "API网关服务" bin\gateway_service.exe

timeout /t 1 /nobreak >nul

echo.
echo 所有服务已启动！
echo.
echo 服务地址:
echo   服务注册中心: http://localhost:8080
echo   用户服务:     http://localhost:8081
echo   订单服务:     http://localhost:8082
echo   API网关:      http://localhost:8083
echo.
echo 查看所有已注册的服务:
echo   curl http://localhost:8080/services
echo.
echo 每个服务都有独立的GUI窗口，可以在窗口中查看实时日志和状态
echo.
pause

