package yamls

task_deploy: {
 type: "k8s",
 component: {
  apiVersion: "apps/v1"
  kind:       "Deployment"
  metadata: name: "flowdeploy"
  spec: {
   selector: matchLabels: app: "flowdeploy"
   replicas: 1
   template: {
    metadata: labels: app: "flowdeploy"
    spec: containers: [{
     name:            "flowdeploy"
     image:           "nginx:1.18-alpine"
     imagePullPolicy: "IfNotPresent"
     ports: [{
      containerPort: 80
     }]
    }]
   }
  }
 }

}