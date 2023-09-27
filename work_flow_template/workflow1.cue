package flowtpls

import (
	  "github.com/workflow/yamls/pods1" // 需要引用的 cue 路径包
)

// 整个工作流模版
workflow: {
      step0:{
      	action: string           // 支持 k8s apply delete 操作，默认不填是 apply 操作
      }
			step1: {
				type: "k8s"              // 工作流种类，目前支持部署 k8s 服务与运行 bash 脚本
				objType: "pod"           // 如果是 k8s 对象，需要指定资源对象，目前仅支持 pod deployment service configmap
				template: pods1.pod1		 // 模版存放位置 yamls/pods1 目录下
				action: step0.action     // step1 会依赖 step0
				status: string           // 因为是 step1 不需要依赖前置状态
			}
		  step2: {
				type: "k8s"
				objType: "configmap"
				template: pods1.cm
				status: step1.status   // step2 会依赖 step1 执行后才能执行
				action: step0.action   // step2 会依赖 step0
     	}
		  step3: {
				type: "k8s"
				objType: "deployment"
				template: pods1.dep
				status: step2.status	 // step3 会依赖 step2 执行后才能执行
				action: step0.action   // step3 会依赖 step0
		  }
}
