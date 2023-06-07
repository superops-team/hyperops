package environment

import (
	"os"

	"github.com/fireworkweb/godotenv"
)

// DefaultEnvStorage 记录所有的环境变量
type DefaultEnvStorage struct{}

// EnvStorage 环境变量存储获取逻辑
type EnvStorage interface {
	Get(string) string
	Set(string, string)
	Load(string) error
	All() []string
	IsTrue(string) bool
}

// NewEnvStorage 创建环境变量存储
func NewEnvStorage() EnvStorage {
	return &DefaultEnvStorage{}
}

// Get 获取指定环境变量
func (es *DefaultEnvStorage) Get(key string) string {
	return os.Getenv(key)
}

// Set 写入环境变量
func (es *DefaultEnvStorage) Set(key string, value string) {
	os.Setenv(key, value)
}

// Load 从文件加载环境变量
func (es *DefaultEnvStorage) Load(filename string) error {
	return godotenv.Load(filename)
}

// All 获取所有环境变量
func (es *DefaultEnvStorage) All() []string {
	return os.Environ()
}

// IsTrue 1 or "true" 均设置为TRUE
func (es *DefaultEnvStorage) IsTrue(key string) bool {
	value := os.Getenv(key)
	return value == "1" || value == "true"
}
