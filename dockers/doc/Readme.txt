安装Helm
  brew install kubernetes-helm

部署Redis
  helm install stable/redis --name single-redis

部署Ingress-nginx
  helm install stable/nginx-ingress --name nginx-ingress

获取版本号
  POD_NAME=$(kubectl get pods -l app.kubernetes.io/name=ingress-nginx -o jsonpath='{.items[0].metadata.name}')
  kubectl exec -it $POD_NAME -- /nginx-ingress-controller --version

部署游戏服务
  kubectl apply -f k8s/depoyments/auth-service.yaml
  kubectl apply -f k8s/depoyments/connect-service.yaml
  kubectl apply -f k8s/depoyments/game-deployment.yaml
  kubectl apply -f k8s/depoyments/world-service.yaml

触发image更新（每次设置不同数字）
  kubectl patch deployment your_deployment -p \
    '{"spec":{"template":{"spec":{"terminationGracePeriodSeconds":31}}}}'
