package flowtpls

import (
	  "github.com/workflow/yamls/pods"
)

workflow: {
      step0:{
      	action: string
      }
			step1: {
				body: pods.pod1
				action: step0.action     // step1会依赖step0
				status: string
			}
			step2: {
				status: step1.status	   // step2会依赖step1的pod状态
				body: pods.pod2
			  action: step0.action     // step2会依赖step0
			}
			step3: {
					status: step2.status	 // step2会依赖step1的pod状态
					body: pods.pod3
					action: step0.action   // step2会依赖step0
		 }
}
