package pods1

cm: {
	 apiVersion: "v1"
    kind:       "ConfigMap"
    metadata: name: "configmap-workflow1"
    data: {
    	aaa: "aaa"
    	bbb: "bbb"
    }
}