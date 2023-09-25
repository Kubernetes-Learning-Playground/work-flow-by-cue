package flowtpls

import (
	  "github.com/workflow/yamls/pods1" // 需要引用的 cue 路径包
)

// 整个工作流模版
workflow: {
      step0:{
      	action: string           // 支持k8s apply delete操作，默认为apply
      }
			step1: {
				type: "k8s"              // 工作流种类，目前支持k8s服务与bash脚本
				objType: "pod"           // 如果是k8s对象，需要指定资源对象，目前仅支持pod deployment service
				template: pods1.pod1		 // 模版存放位置
				action: step0.action     // step1会依赖step0
				status: string
			}
		  step2: {
				type: "k8s"
				objType: "configmap"
				template: pods1.cm
				status: step2.status	 // step2会依赖step1的pod状态
				action: step0.action   // step2会依赖step0
     	}
		  step3: {
				type: "k8s"
				objType: "deployment"
				template: pods1.dep
				status: step3.status	 // step2会依赖step1的pod状态
				action: step0.action   // step2会依赖step0
		  }
}
