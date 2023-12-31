## workflow 工作流

### 项目思路与功能
项目背景：一般在集群内都会有部署需求，因此需要类似工作流的服务。(需要顺序执行的业务场景，ex: 部署时 b 依赖 a，c 依赖 b )

支持功能：
1. 可提供 部署 k8s 服务与运行 bash 脚本执行功能
2. 目前k8s支持 **pods** **deployment** **service** **configmaps** 资源，并支持 apply delete 操作
3. 支持多 step "串行"执行功能( cue 模版中 step 字段中设置 status: stepx.status，就会等待前一个 step 执行后再执行)
![](https://github.com/Kubernetes-Learning-Playground/work-flow-by-cue/blob/main/image/%E6%97%A0%E6%A0%87%E9%A2%98-2023-08-10-2343.png?raw=true)

- 目前支持 cue 文件配置工作流，后续会对其进行适配或扩展，让工作流支持 yaml 文件

更多配置文件示例可[参考](./work_flow_template)
```cue
// 整个工作流模版
workflow: {
      step0:{
      	action: string             // 支持k8s apply delete操作，莫认为apply
      }
      step1: {
        type: "k8s"                // 工作流种类，目前支持k8s服务与bash脚本
        objType: "pod"             // 如果是k8s对象，需要指定资源对象，目前仅支持pod deployment service
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
        status: step2.status	   // step3会依赖step2的pod状态
      }
      step4: {
        type: "k8s"
        objType: "pod"
        template: pods.pod3
        status: step3.status	   // step4会依赖step3的pod状态
        action: step0.action       // step4会依赖step0
      }
      step5: {
        type: "k8s"
        objType: "deployment"
        template: pods.dep
        status: step4.status	 // step5会依赖step4的pod状态
        action: step0.action     // step5会依赖step0
      }
	  step6: {
        type: "k8s"
        objType: "service"
        template: pods.svc
        status: step4.status	 // step6会依赖step4的pod状态
        action: step0.action     // step6会依赖step0
      }

}
```
更多配置文件示例可[参考](./yamls)
- bash 脚本 cue 模版
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

- pod cue 模版
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

### 接口调用实现：
- 注册工作流接口
![](https://github.com/Kubernetes-Learning-Playground/work-flow-by-cue/blob/main/image/img.png?raw=true)
- 执行工作流接口
![](https://github.com/Kubernetes-Learning-Playground/work-flow-by-cue/blob/main/image/img_1.png?raw=true)



