package pods

cm: {
	 apiVersion: "v1"
    kind:       "ConfigMap"
    metadata: name: "test-demo"
    data: {
    	aaa: "aaa"
    	bbb: "bbb"

    }
}