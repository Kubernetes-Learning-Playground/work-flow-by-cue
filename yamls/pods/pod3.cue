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
				command: ["sh", "-c", "echo \"pod3-step3\" && sleep 3600"]
			}]
		}
}