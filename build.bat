@echo off
chcp 65001 >nul
echo 开始编译微服务...

REM 编译服务注册中心
echo 编译服务注册中心...
cd center_service
go build -o ..\bin\center_service.exe main.go
if %errorlevel% equ 0 (
    echo ✓ 服务注册中心编译成功: bin\center_service.exe
) else (
    echo ✗ 服务注册中心编译失败
    exit /b 1
)

REM 编译用户服务
echo 编译用户服务...
cd ..\user_service
go build -o ..\bin\user_service.exe main.go
if %errorlevel% equ 0 (
    echo ✓ 用户服务编译成功: bin\user_service.exe
) else (
    echo ✗ 用户服务编译失败
    exit /b 1
)

REM 编译订单服务
echo 编译订单服务...
cd ..\order_service
go build -o ..\bin\order_service.exe main.go
if %errorlevel% equ 0 (
    echo ✓ 订单服务编译成功: bin\order_service.exe
) else (
    echo ✗ 订单服务编译失败
    exit /b 1
)

REM 编译网关服务
echo 编译网关服务...
cd ..\gateway_service
go build -o ..\bin\gateway_service.exe main.go
if %errorlevel% equ 0 (
    echo ✓ 网关服务编译成功: bin\gateway_service.exe
) else (
    echo ✗ 网关服务编译失败
    exit /b 1
)

REM 编译测试客户端
echo 编译测试客户端...
cd ..\client
go build -o ..\bin\client.exe main.go
if %errorlevel% equ 0 (
    echo ✓ 测试客户端编译成功: bin\client.exe
) else (
    echo ✗ 测试客户端编译失败
    exit /b 1
)

cd ..
echo.
echo 所有服务编译完成！
echo 可执行文件位置:
echo   - bin\center_service.exe    (服务注册中心，端口8080)
echo   - bin\user_service.exe      (用户服务，端口8081)
echo   - bin\order_service.exe     (订单服务，端口8082)
echo   - bin\gateway_service.exe   (API网关服务，端口8083)
echo   - bin\client.exe            (API测试客户端)
echo.
echo 运行方式:
echo   双击运行或在命令行执行:
echo     bin\center_service.exe
echo     bin\user_service.exe
echo     bin\order_service.exe
echo     bin\gateway_service.exe
echo     bin\client.exe
echo.
echo 或使用 start.bat 一键启动所有服务

