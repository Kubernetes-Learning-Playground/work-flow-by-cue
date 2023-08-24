package yamls


task_svc: {
 type: "k8s",  // k8s服务相关
 component: {
  apiVersion: "v1"
  kind:       "Service"
  metadata: name: "flowsvc"
  spec: {
   type: "ClusterIP"
   ports: [{
    port:       80
    targetPort: 80
   }]
   selector: {//service通过selector和pod建立关联
    app: "flowdeploy"
   }
  }
 }
}