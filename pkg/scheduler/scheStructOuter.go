package scheduler

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	"strconv"
	"time"
	"zbx.io/externalscheduler/pkg/util"
)

type Scheduler struct {
	nodeManager
	podManager

	stopCh       chan struct{}
	kubeClient   kubernetes.Interface // 存储k8s 连接客户端。该值将在scheduler start时赋值。 informer也要用到
	podLister    listerscorev1.PodLister
	nodeLister   listerscorev1.NodeLister
	cachedstatus map[string]*NodeUsage
}

func NewScheduler() *Scheduler {
	klog.Infof("New Scheduler")
	s := &Scheduler{
		stopCh:       make(chan struct{}),
		cachedstatus: make(map[string]*NodeUsage),
	}
	s.nodeManager.init()
	s.podManager.init()
	return s
}

func (s *Scheduler) Start() {
	kubeClient, err := util.NewClient()
	util.Check(err)
	s.kubeClient = kubeClient
	/*
		1、建立一个informer工厂类
		2、用工厂.api资源.informer() 创建某个资源的informer。 informer用来注册回调函数，watch到资源变更会触发回调
		3、用工厂.api资源.Lister()  创建某个资源的lister.  lister的作用主要是用来查询该资源
		4、informer工厂实例.Start 将会初始化所有已请求的informers。 WaitForCacheSync 等待所有已经启动的informer的Cache同步完成。

	*/

	// 作用： 建立一个informer工厂类。
	// 参数1：由config生成的client。生成工厂需要这个client。
	// 参数2: 指的是每过一段时间，清空本地缓存，从apiserver做一次list。 注意在大规模集群中，list的代价不容小视。
	// 参数3： 不传：监控所有namespace中特定资源对象的工厂实例。可传指定namespace、指定过滤方式等。
	// 其他： 同样的，如果上面是crd资源的informer，则为 crdInformerFactory := crdinformers.NewSharedInformerFactory(client2, time.Second*30) 。其中crdinformers来自 crdinformers "k8s.io/sample-controller/pkg/generated/informers/externalversions"
	informerFactory := informers.NewSharedInformerFactoryWithOptions(s.kubeClient, time.Hour*1)

	s.podLister = informerFactory.Core().V1().Pods().Lister()
	s.nodeLister = informerFactory.Core().V1().Nodes().Lister()

	informer := informerFactory.Core().V1().Pods().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.onAddPod,
		UpdateFunc: s.onUpdatePod,
		DeleteFunc: s.onDelPod,
	})

	informerFactory.Start(s.stopCh)
	informerFactory.WaitForCacheSync(s.stopCh)
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) Bind(args extenderv1.ExtenderBindingArgs) (*extenderv1.ExtenderBindingResult, error) {
	klog.InfoS("Bind", "pod", args.PodName, "namespace", args.PodNamespace, "podUID", args.PodUID, "node", args.Node)
	var err error
	var res *extenderv1.ExtenderBindingResult
	binding := &corev1.Binding{
		ObjectMeta: metav1.ObjectMeta{Name: args.PodName, UID: args.PodUID},
		Target:     corev1.ObjectReference{Kind: "Node", Name: args.Node},
	}
	current, err := s.kubeClient.CoreV1().Pods(args.PodNamespace).Get(context.Background(), args.PodName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "Get pod failed")
	}
	err = util.LockNode(args.Node)
	if err != nil {
		klog.ErrorS(err, "Failed to lock node", "node", args.Node)
	}
	//defer util.ReleaseNodeLock(args.Node)

	tmppatch := make(map[string]string)
	tmppatch[util.DeviceBindPhase] = "allocating"
	tmppatch[util.BindTimeAnnotations] = strconv.FormatInt(time.Now().Unix(), 10)

	err = util.PatchPodAnnotations(current, tmppatch)
	if err != nil {
		klog.ErrorS(err, "patch pod annotation failed")
	}
	if err = s.kubeClient.CoreV1().Pods(args.PodNamespace).Bind(context.Background(), binding, metav1.CreateOptions{}); err != nil {
		klog.ErrorS(err, "Failed to bind pod", "pod", args.PodName, "namespace", args.PodNamespace, "podUID", args.PodUID, "node", args.Node)
	}
	if err == nil {
		res = &extenderv1.ExtenderBindingResult{
			Error: "",
		}
	} else {
		res = &extenderv1.ExtenderBindingResult{
			Error: err.Error(),
		}
	}
	klog.Infoln("After Binding Process")
	return res, nil
}