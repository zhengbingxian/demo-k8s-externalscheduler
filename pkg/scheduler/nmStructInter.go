package scheduler

import "sync"

type nodeManager struct {
	nodes map[string]*NodeInfo
	mutex sync.Mutex
}

// 初始化方法，只是将nodes字段map初始化一下
func (m *nodeManager) init() {
	m.nodes = make(map[string]*NodeInfo)
}
