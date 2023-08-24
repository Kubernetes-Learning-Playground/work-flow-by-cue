//user: {
//	name: "jiang",
//	age: 18
//}
//
// pod: {
//  "apiVersion": 123,
//  "kind": "Pod",
//  "metadata": {
//    "name": "mypod"
//  },
//  "spec": {
//    "containers": [
//      {
//        "image": "nginx:1.18-alpine"
//      }
//    ]
//  }
//
//}


#metadata:{
   name: string
   namespace: string | *"default"
}

// 需要赋予验证规则
metadata: #metadata & {
 "name": "nginx"
}

pod: {
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": metadata, // 拆分对象，可以引用metadata
  "spec": {
    "containers": [
      {
        "image": "nginx:1.18-alpine"
      }
    ]
  }

}