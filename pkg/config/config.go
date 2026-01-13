package config

import (
	"os"
	"strconv"
)

// GetEnv 获取环境变量，如果不存在返回默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt 获取整数环境变量
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetServiceAddress 获取服务地址（用于服务注册）
// 在Docker环境中返回Docker服务名，否则返回localhost
func GetServiceAddress(serviceName string, defaultAddress string) string {
	// 优先使用环境变量
	if addr := os.Getenv(serviceName + "_ADDRESS"); addr != "" {
		return addr
	}
	
	// 检查是否在Docker环境中（通过环境变量ENV或容器名）
	if os.Getenv("ENV") != "" || os.Getenv("HOSTNAME") != "" {
		// Docker环境中使用服务名
		return serviceName
	}
	
	// 默认值
	return defaultAddress
}

