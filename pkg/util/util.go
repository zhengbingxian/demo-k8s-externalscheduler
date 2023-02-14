/*
一些共用的方法
*/

package util

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NewClient 尝试获取config，并使用该配置与k8s建立连接。 如果连接成功则返回k8s连接客户端。
func NewClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig() // 首先尝试从pod内获取config
	if err != nil {
		// 如果没有，则从环境变量KUBECONFIG 或 .kube/config获取config
		kubeConfig := os.Getenv("KUBECONFIG")
		if kubeConfig == "" {
			kubeConfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			return nil, err
		}
	}
	// k8s原生资源，生成client的方式：使用config构建k8s连接客户端
	// 如果是crd资源就是用client2, err := clientset.NewForConfig(cfg)，其中clientset "k8s.io/sample-controller/pkg/generated/clientset/versioned"
	client, err := kubernetes.NewForConfig(config)
	return client, err
}

// Check 检查错误并报告
func Check(err error) {
	if err != nil {
		klog.Fatal(err)
	}
}

type ContainerDevice struct {
	UUID string
	Type string
}

type ContainerDevices []ContainerDevice

type PodDevices []ContainerDevices

func DecodePodDevices(str string) PodDevices {
	if len(str) == 0 {
		return PodDevices{}
	}
	var pd PodDevices
	for _, s := range strings.Split(str, ";") {
		cd := DecodeContainerDevices(s)
		pd = append(pd, cd)
	}
	return pd
}

func DecodeContainerDevices(str string) ContainerDevices {
	if len(str) == 0 {
		return ContainerDevices{}
	}
	cd := strings.Split(str, ":")
	contdev := ContainerDevices{}
	tmpdev := ContainerDevice{}
	//fmt.Println("before container device", str)
	if len(str) == 0 {
		return contdev
	}
	for _, val := range cd {
		if strings.Contains(val, ",") {
			//fmt.Println("cd is ", val)
			tmpstr := strings.Split(val, ",")
			tmpdev.UUID = tmpstr[0]
			tmpdev.Type = tmpstr[1]
			contdev = append(contdev, tmpdev)
		}
	}
	//fmt.Println("Decoded container device", contdev)
	return contdev
}

var kubeClient kubernetes.Interface

func init() {
	kubeClient, _ = NewClient()
}

func LockNode(nodeName string) error {
	ctx := context.Background()
	node, err := kubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if _, ok := node.ObjectMeta.Annotations[NodeLockTime]; !ok {
		return SetNodeLock(nodeName)
	}
	lockTime, err := time.Parse(time.RFC3339, node.ObjectMeta.Annotations[NodeLockTime])
	if err != nil {
		return err
	}
	if time.Since(lockTime) > time.Minute*5 {
		klog.InfoS("Node lock expired", "node", nodeName, "lockTime", lockTime)
		err = ReleaseNodeLock(nodeName)
		if err != nil {
			klog.ErrorS(err, "Failed to release node lock", "node", nodeName)
			return err
		}
		return SetNodeLock(nodeName)
	}
	return fmt.Errorf("node %s has been locked within 5 minutes", nodeName)
}
