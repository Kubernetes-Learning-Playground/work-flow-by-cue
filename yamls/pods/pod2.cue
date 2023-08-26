package pods

pod2: {
	  apiVersion: "v1"
		kind:       "Pod"
		metadata: name: "pod2"
		spec: {
			restartPolicy: "Never"
			initContainers: [{
					name:  "init-myservice"
					image: "busybox:1.28"
					command: ["sh", "-c", "sleep 10"]
		  }]
			containers: [{
				name:  "step1"
				image: "busybox:1.28"
				command: ["sh", "-c", "echo \"pod2-step1\" && sleep 3600"]
			}]
		}
}