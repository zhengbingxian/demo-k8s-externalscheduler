/*
 * Copyright © 2021 peizhaoyou <peizhaoyou@4paradigm.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scheduler

import (
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"zbx.io/externalscheduler/pkg/util"
)

// 就是把podUid ~ pod信息，加入到map里维护起来
func (m *podManager) addPod(pod *corev1.Pod, nodeID string, devices util.PodDevices) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, ok := m.pods[pod.UID] // 如果在map里没有找到，则说明需要将该pod放到map里
	if !ok {
		pi := &podInfo{Name: pod.Name, Uid: pod.UID}
		m.pods[pod.UID] = pi
		pi.Namespace = pod.Namespace
		pi.Name = pod.Name
		pi.Uid = pod.UID
		pi.NodeID = nodeID
		pi.Devices = devices
		klog.Info(pod.Name + "Added")
	}
}

func (m *podManager) delPod(pod *corev1.Pod) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	pi, ok := m.pods[pod.UID]
	if ok {
		klog.Infof(pi.Name + " deleted")
		delete(m.pods, pod.UID)
	}
}

func (m *podManager) GetScheduledPods() (map[k8stypes.UID]*podInfo, error) {
	return m.pods, nil
}
