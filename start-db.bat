@echo off
chcp 65001 >nul

echo 启动微服务数据库...

REM 检查Docker是否运行
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: Docker未运行，请先启动Docker
    pause
    exit /b 1
)

REM 启动数据库
docker-compose up -d

echo.
echo 等待数据库启动...
timeout /t 5 /nobreak >nul

REM 检查数据库状态
echo.
echo 数据库状态:
docker-compose ps

echo.
echo 数据库连接信息:
echo   用户服务:
echo     MySQL: localhost:3306 (user_service/user_pass)
echo     Redis: localhost:6379
echo.
echo   订单服务:
echo     MySQL: localhost:3307 (order_service/order_pass)
echo     Redis: localhost:6380
echo.
echo   注册中心:
echo     MySQL: localhost:3308 (registry_service/registry_pass)
echo.
echo 查看日志: docker-compose logs -f
echo 停止数据库: docker-compose stop
echo 删除数据库: docker-compose down
echo.
pause

