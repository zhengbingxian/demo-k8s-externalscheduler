package scheduler

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"zbx.io/externalscheduler/pkg/util"
)

// 解析pod信息， 然后就是把podUid ~ pod信息，加入到map里维护起来
func (s *Scheduler) onAddPod(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		klog.Errorf("unknown add object type")
		return
	}
	nodeID, ok := pod.Annotations["zbx.com/mockgpu-node"] // 此处nodeID是什么？
	if !ok {
		return
	}
	ids, ok := pod.Annotations["zbx.com/myid"] //myid是node的id还是podid还是vgpu编号？
	if !ok {
		return
	}
	// 如果已经failed或succeed，则删除该pod。
	if IsPodInTerminatedState(pod) {
		s.delPod(pod)
		return
	}
	podDev := util.DecodePodDevices(ids)
	s.addPod(pod, nodeID, podDev)
}

func IsPodInTerminatedState(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded
}
