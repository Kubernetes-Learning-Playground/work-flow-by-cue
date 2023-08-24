
#metadata:{
   name: string
   namespace: string | *"default"
}
metadata: #metadata & {
 "name": "nginx"
}
#container: {
    name?: string
    image: string
}
#containers: [...#container]
containers: #containers & [
      {
        "name": "nginx"
        "image": "nginx:1.18-alpine"
      }

]
param: {
   "apiVersion":"v1",
   "kind": "Pod",
    name: "abc"
}

pod: {
  "apiVersion": param.apiVersion,
  "kind": param.kind,
  "metadata":metadata,
  "spec": {
    "containers": containers
  }

}