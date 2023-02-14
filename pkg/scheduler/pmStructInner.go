package scheduler

import (
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sync"
)

type podManager struct {
	pods  map[k8stypes.UID]*podInfo
	mutex sync.Mutex
}

// 初始化方法，只是将pods字段map初始化一下
func (m *podManager) init() {
	m.pods = make(map[k8stypes.UID]*podInfo)
}
