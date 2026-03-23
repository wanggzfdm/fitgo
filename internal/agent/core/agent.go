package core

import (
	"context"
	"fmt"
)

// Agent 定义智能体接口
type Agent interface {
	ID() string
	Name() string
	Description() string
	Handle(ctx context.Context, input interface{}) (interface{}, error)
}

// BaseAgent 提供智能体的基础实现
type BaseAgent struct {
	id          string
	name        string
	description string
}

// NewBaseAgent 创建基础智能体
func NewBaseAgent(id, name, description string) *BaseAgent {
	return &BaseAgent{
		id:          id,
		name:        name,
		description: description,
	}
}

func (a *BaseAgent) ID() string {
	return a.id
}

func (a *BaseAgent) Name() string {
	return a.name
}

func (a *BaseAgent) Description() string {
	return a.description
}

// Handle 默认的消息处理函数，子类应该重写这个方法
func (a *BaseAgent) Handle(ctx context.Context, input interface{}) (interface{}, error) {
	return nil, fmt.Errorf("handle method not implemented")
}
