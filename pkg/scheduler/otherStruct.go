package scheduler

import (
	k8stypes "k8s.io/apimachinery/pkg/types"
	"zbx.io/externalscheduler/pkg/util"
)

type podInfo struct {
	Namespace string
	Name      string
	Uid       k8stypes.UID
	NodeID    string
	Devices   util.PodDevices
	CtrIDs    []string
}

type NodeInfo struct {
	ID      string
	Devices []DeviceInfo
}

type NodeUsage struct {
	Devices []*DeviceUsage
}

type DeviceUsage struct {
	Id        string
	Used      int32
	Count     int32
	Usedmem   int32
	Totalmem  int32
	Usedcores int32
	Type      string
	Health    bool
}
