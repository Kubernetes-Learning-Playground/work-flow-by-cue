## workflow工作流

### 项目思路与功能
项目背景：一般在集群内都会有类似工作流(需要顺序执行的业务场景，ex: 部署时b依赖a，c依赖b)

支持功能：
1. 可提供 k8s部署与bash脚本执行功能
2. 目前k8s仅支持pods资源，并支持apply delete操作
3. 支持多step"串行"或"并行"执行功能(cue模版中step字段中设置status: stepx.status，就会等待前一个step执行后再执行)
![](https://github.com/Kubernetes-Learning-Playground/work-flow-by-cue/blob/main/image/%E6%97%A0%E6%A0%87%E9%A2%98-2023-08-10-2343.png?raw=true)


```cue
// 整个工作流模版
workflow: {
      step0:{
      	action: string             // 支持k8s apply delete操作，莫认为apply
      }
      step1: {
        type: "k8s"                // 工作流种类，目前支持k8s服务与bash脚本
        objType: "pod"             // 如果是k8s对象，需要指定资源对象，目前仅支持pod
        template: pods.pod         // 模版存放位置
        action: step0.action       // step1会依赖step0
        status: string
      }
      step2: {
        type: "k8s"
        objType: "pod"
        template: pods.pod2
        status: step1.status	   // step2会依赖step1的pod状态
        action: step0.action       // step2会依赖step0
      }
      step3: {
        type: "bash",
        script: pods.script
        status: step2.status	   // step2会依赖step1的pod状态
      }
     step4: {
        type: "k8s"
        objType: "pod"
        template: pods.pod3
        status: step3.status	   // step2会依赖step1的pod状态
        action: step0.action       // step2会依赖step0
     }

}

```

- bash脚本cue模版
```cue
package pods

// 脚本内容
script: """
#!/bin/bash

for i in {1..10}; do
echo $i
sleep 1
done
"""
```

- pod cue模版
```cue
package pods

pod3: {
  apiVersion: "v1"
    kind:       "Pod"
    metadata: name: "pod3"
    spec: {
        restartPolicy: "Never"
        containers: [{
            name:  "step1"
            image: "busybox:1.28"
            command: ["sh", "-c", "echo \"pod2-step1\" && sleep 3600"]
        }]
    }
}

```



