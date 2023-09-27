# 辅助脚本，删除创建好的资源
kubectl delete pods pod1 pod2 pod3 pod-workflow1
kubectl delete cm test-demo configmap-workflow1
kubectl delete deploy flowdeploy deployment-workflow1
kubectl delete svc flowsvc