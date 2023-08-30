package pods

cm: {
	 apiVersion: "v1"
    kind:       "ConfigMap"
    metadata: name: "test-demo"
    data: {
    	aaa: "aaa"
    	bbb: "bbb"

    	test_properties: """
											 aaa=aaa
											 bbb=5
											"""

    	test_properties1: """
											  aaa=aaa
											  bbb=5
												"""

    }
}