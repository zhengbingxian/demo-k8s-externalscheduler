# 最简单的调度器
调度器最简单的方式可以直接用shell来简单的实现！
1、在k8s集群里找到pod指定调度器名称的，且node字段为空的。
2、在k8s集群里找到一个node。
3、向k8s apiserver发送post请求，请求该pod绑定node即可

## 使用说明

```shell
# 1、使用一个terminal启动proxy
$ kubectl proxy  

# 2、打开第二个terminal，创建pod，指定调度器
$ kubectl apply -f podUseMyScheduler.yaml

# 3、运行scheduler, 发现已经将该pod调度到node上
$ ./scheduler.sh
{
  "kind": "Status",
  "apiVersion": "v1",
  "metadata": {
    
  },
  "status": "Success",
  "code": 201
}Assigned nginx to node1    
```