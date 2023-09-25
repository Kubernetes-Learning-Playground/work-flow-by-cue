package flowtpls

import (
	  "github.com/workflow/yamls/pods" // 需要引用的 cue 路径包
)

// 整个工作流模版
workflow: {
      step0:{
      	action: string           // 支持k8s apply delete操作，默认为apply
      }
			step1: {
				type: "k8s"              // 工作流种类，目前支持k8s服务与bash脚本
				objType: "pod"           // 如果是k8s对象，需要指定资源对象，目前仅支持pod deployment service
				template: pods.pod1		   // 模版存放位置
				action: step0.action     // step1会依赖step0
				status: string
			}
			step2: {
				type: "k8s"
				objType: "pod"
				template: pods.pod2
				status: step1.status	   // step2会依赖step1的pod状态
			  action: step0.action     // step2会依赖step0
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
				status: step3.status	 // step2会依赖step1的pod状态
				action: step0.action   // step2会依赖step0
		  }
		  step5: {
				type: "k8s"
				objType: "configmap"
				template: pods.cm
				status: step4.status	 // step2会依赖step1的pod状态
				action: step0.action   // step2会依赖step0
     	}
		  step6: {
				type: "k8s"
				objType: "deployment"
				template: pods.dep
				status: step5.status	 // step2会依赖step1的pod状态
				action: step0.action   // step2会依赖step0
		  }
		  step7: {
				type: "k8s"
				objType: "service"
				template: pods.svc
				status: step6.status	 // step2会依赖step1的pod状态
				action: step0.action   // step2会依赖step0
		  }
}
