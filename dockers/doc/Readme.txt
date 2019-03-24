安装Helm
  brew install kubernetes-helm

部署Redis
  kubectl apply -f k8s/deps/redis/redis-cluster.yaml
  kubectl exec -it redis-cluster-0 -- redis-cli --cluster create --cluster-replicas 1 \
  $(kubectl get pods -l app=redis-cluster -o jsonpath='{range.items[*]}{.status.podIP}:6379 ')

部署Mysql
  helm install stable/mysql --name single-mysql

部署Ingress-nginx
  helm install stable/nginx-ingress --name nginx-ingress

获取版本号
  POD_NAME=$(kubectl get pods -l app.kubernetes.io/name=ingress-nginx -o jsonpath='{.items[0].metadata.name}')
  kubectl exec -it $POD_NAME -- /nginx-ingress-controller --version

部署ServicePerPod
  # Create metacontroller namespace.
  kubectl create namespace metacontroller
  # Create metacontroller service account and role/binding.
  kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/metacontroller/master/manifests/metacontroller-rbac.yaml
  # Create CRDs for Metacontroller APIs, and the Metacontroller StatefulSet.
  kubectl apply -f https://raw.githubusercontent.com/GoogleCloudPlatform/metacontroller/master/manifests/metacontroller.yaml  

  kubectl create configmap service-per-pod-hooks -n metacontroller --from-file=k8s/deps/service-per-pod/hooks
  kubectl apply -f k8s/deps/service-per-pod/service-per-pod.yaml
  

部署游戏服务
  kubectl apply -f k8s/deployments/game-deployment.yaml
  kubectl apply -f k8s/deployments/world-service.yaml

触发image更新（每次设置不同数字）
  kubectl patch statefulsets.apps gos-game-app -p '{"spec":{"template":{"spec":{"terminationGracePeriodSeconds":18}}}}'
  kubectl patch deployment gos-world-app -p '{"spec":{"template":{"spec":{"terminationGracePeriodSeconds":18}}}}'
