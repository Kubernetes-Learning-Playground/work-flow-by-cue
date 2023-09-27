package flowtpls

import (
	  "github.com/workflow/yamls/pods" // 需要引用的 cue 路径包
)

// 整个工作流模版
workflow: {
      step0: {
      	action: "apply"           // 支持 k8s apply delete 操作，默认不填是 apply 操作
      }
			step1: {
				type: "k8s"              // 工作流种类，目前支持部署 k8s 服务与运行 bash 脚本
				objType: "pod"           // 如果是 k8s 对象，需要指定资源对象，目前仅支持 pod deployment service configmap
				template: pods.pod1		   // 模版存放位置 yamls/pods 目录下
				action: step0.action     // step1 会依赖 step0
				status: string
			}
			step2: {
				type: "k8s"
				objType: "pod"
				template: pods.pod2
				status: step1.status	   // step2 会依赖 step1 执行结束后再执行
			  action: step0.action     // step2 会依赖 step0
			}
			step3: {
				type: "bash",
				script: pods.script
				status: step2.status	   // step2 会依赖 step1 执行结束后再执行
			}
			step4: {
				type: "k8s"
				objType: "pod"
				template: pods.pod3
				status: step3.status	 // step4 会依赖 step3 执行结束后再执行
				action: step0.action   // step4 会依赖 step0
		  }
		  step5: {
				type: "k8s"
				objType: "configmap"
				template: pods.cm
				status: step4.status	 // step5 会依赖 step4 执行结束后再执行
				action: step0.action   // step5 会依赖 step0
     	}
		  step6: {
				type: "k8s"
				objType: "deployment"
				template: pods.dep
				status: step5.status	 // step6 会依赖 step5 执行结束后再执行
				action: step0.action   // step6 会依赖 step0
		  }
		  step7: {
				type: "k8s"
				objType: "service"
				template: pods.svc
				status: step6.status	 // step7 会依赖 step6 执行结束后再执行
				action: step0.action   // step7 会依赖 step0
		  }
}
