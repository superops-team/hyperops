package context

import (
	"fmt"
	"strings"
	"sync"

	"go.starlark.net/starlark"
)

var (
	once     sync.Once
	instance *secretsManager
)

// secretsManager 密码管理器,管理所有安全相关的秘钥信息
type secretsManager struct {
	sync.Mutex
	allSecrets map[string]struct{} // Set类型
}

// NewSecretsManager 返回密码管理器对象,该对象记录了所有的密码信息
func NewSecretsManager() *secretsManager {
	once.Do(func() {
		instance = &secretsManager{
			allSecrets: make(map[string]struct{}),
		}
	})
	return instance
}

// AddSecret 添加密码
func (s *secretsManager) AddSecret(secret string) {
	s.Lock()
	defer s.Unlock()
	s.allSecrets[secret] = struct{}{}
}

// HasSecret 是否包含指定密码
func (s *secretsManager) HasSecret(secret string) bool {
	s.Lock()
	defer s.Unlock()
	_, ok := s.allSecrets[secret]
	return ok
}

// SafeReplace 安全替换
func (s *secretsManager) SafeReplace(msg string) string {
	s.Lock()
	defer s.Unlock()
	newMsg := msg
	// thread safe range
	for secret := range s.allSecrets {
		dst := strings.Repeat("*", len(secret))
		newMsg = strings.Replace(newMsg, secret, dst, -1)
	}
	return newMsg
}

// SafePrint for replace builtin "print"
func SafePrint(thread *starlark.Thread, msg string) {
	sm := NewSecretsManager()
	safeMsg := sm.SafeReplace(msg)
	fmt.Println(safeMsg)
}
