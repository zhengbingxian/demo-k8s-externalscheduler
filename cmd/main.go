package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"net/http"
	"zbx.io/externalscheduler/pkg/config"
	"zbx.io/externalscheduler/pkg/scheduler"
	"zbx.io/externalscheduler/pkg/scheduler/routes"
)

var (
	rootCmd = &cobra.Command{
		Use:   "externalscheduler",
		Short: "kubernetes external scheduler",
		Run: func(cmd *cobra.Command, args []string) {
			start()
		},
	}
)

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false
	rootCmd.Flags().StringVar(&config.GrpcBind, "grpc_bind", "127.0.0.1:9090", "grpc server bind address")
	rootCmd.Flags().StringVar(&config.HttpBind, "http_bind", "127.0.0.1:8080", "http server bind address")
	rootCmd.Flags().StringVar(&config.TlsCertFile, "cert_file", "", "tls cert file")
	rootCmd.Flags().StringVar(&config.TlsKeyFile, "key_file", "", "tls key file")
	rootCmd.Flags().StringVar(&config.SchedulerName, "scheduler-name", "", "the name to be added to pod.spec.schedulerName if not empty")
}

func start() {
	// 这是主要是用高效的informer，来收集k8s集群信息，为后续信息填充做准备。 为本scheduler特有。
	// 最低端的、无视性能的方法，就是直接用client查k8s的信息。 也是可以实现效果的
	sch := scheduler.NewScheduler()
	sch.Start()
	defer sch.Stop()

	// start http server. github.com/julienschmidt/httprouter为轻量级高性能http请求路由器
	// 只需三步，new，配置路由及方法，启动
	router := httprouter.New()
	router.POST("/filter", routes.PredicateRoute(sch))
	router.POST("/bind", routes.Bind(sch))
	router.POST("/webhook", routes.WebHookRoute())
	klog.Info("listen on ", config.HttpBind)
	if len(config.TlsCertFile) == 0 || len(config.TlsKeyFile) == 0 {
		if err := http.ListenAndServe(config.HttpBind, router); err != nil {
			klog.Fatal("Listen and Serve error, ", err)
		}
	} else {
		if err := http.ListenAndServeTLS(config.HttpBind, config.TlsCertFile, config.TlsKeyFile, router); err != nil {
			klog.Fatal("Listen and Serve error, ", err)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		klog.Fatal(err)
	}
}
