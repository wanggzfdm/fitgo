package manager

import (
	"context"
	"fmt"
	"sync"

	"fitgo/internal/agent/core"
)

// Manager 管理智能体生命周期
type Manager struct {
	agents map[string]core.Agent
	mu     sync.RWMutex
}

// New 创建新的智能体管理器
func New() *Manager {
	return &Manager{agents: make(map[string]core.Agent)}
}

// Register 注册智能体
func (m *Manager) Register(agent core.Agent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.agents[agent.ID()]; ok {
		return fmt.Errorf("agent %s already exists", agent.ID())
	}
	m.agents[agent.ID()] = agent
	return nil
}

// Get 获取智能体
func (m *Manager) Get(id string) (core.Agent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.agents[id]
	return a, ok
}

// Process 调用指定智能体处理输入
func (m *Manager) Process(ctx context.Context, id string, input interface{}) (interface{}, error) {
	a, ok := m.Get(id)
	if !ok {
		return nil, fmt.Errorf("agent %s not found", id)
	}
	return a.Handle(ctx, input)
}
