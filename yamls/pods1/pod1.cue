package pods1
pod1: {
	  apiVersion: "v1"
		kind:       "Pod"
		metadata: name: "pod-workflow1"
		spec: {
			restartPolicy: "Never"
			initContainers: [{
				name:  "init-myservice"
				image: "busybox:1.28"
				command: ["sh", "-c", "sleep 5"]
			}]
			containers: [{
				name:  "step1"
				image: "busybox:1.28"
				command: ["sh", "-c", "echo \"pod1-step1\" && sleep 3600"]
			}]
		}
}
