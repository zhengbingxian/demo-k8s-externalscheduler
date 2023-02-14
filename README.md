# demo-k8s-externalscheduler

## 目的
为初学者了解external scheduler的原理。
* 1TheSimpleScheduler 为最简单的shell实现的扩展调度器，可以体验一下，就知道扩展调度器是干啥的了
* todo 此go项目，暂不可运行，暂不继续开发。

## 背景知识
每次调度，可以成为一个调度上下文。 分为两个阶段：
* 调度周期  Scheduling Cycle  为pod选择一个节点，SC是串行的
* 绑定周期  Binding Cycle， 将sc的决策结果应用到集群中，可以并行。

在整个调度上下文中，定义了好几个可以扩展的点。比如实现要扩展bind过程，
就配置下bind动作，接口为/bind，在扩展调度器中实现/bind对应的方法即可。所有扩展点介绍随便找了一下[见此](https://dev.to/cylon/kube-schedulerde-diao-du-shang-xia-wen-3eik#podsc)





### scheduler可进行策略配置
以下为一些示例配置，主要是三个字段：
* predicates 预选，过滤不符合运行条件的node。  可以定义一些规则
* priorities  优选，对node进行打分。   这里可以额外配置一些权重。
* extenders   保存扩展调度器的一些动作、接口等。 比如filterVerb: filter表示访问扩展调度器的urlPrefix+filter接口，即可触发
  * 其中filterVerb， 被调用时返回节点列表 extenderFilterResult。 所以我们可以额外方法里对这些节点列表进行裁剪。
  *  "prioritize"返回节点的优先级(schedulerapi.HostPriorityList). 
  * bind用于绑定pod到node。这个过程可以交给扩展调度器自行实现。

```json
{
  "kind" : "Policy",
  "apiVersion" : "v1",
  "predicates" : [      
    {"name" : "PodFitsHostPorts"},
    {"name" : "PodFitsResources"},
    {"name" : "NoDiskConflict"},
    {"name" : "MatchNodeSelector"},
    {"name" : "HostName"}
  ],
  "priorities" : [
    {"name" : "LeastRequestedPriority", "weight" : 1},
    {"name" : "BalancedResourceAllocation", "weight" : 1},
    {"name" : "ServiceSpreadingPriority", "weight" : 1},
    {"name" : "EqualPriority", "weight" : 1}
  ],
  "extenders": [
    {
      "urlPrefix": "https://127.0.0.1:443",
      "filterVerb": "filter",   
      "bindVerb": "bind"
    }
  ]
}



```



配置pod使用此调度器进行调度
```shell
cat > podUseExternalScheduler.yaml << EOF
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  schedulerName: my-scheduler
  containers:
  - name: nginx
    image: nginx:1.10

```



