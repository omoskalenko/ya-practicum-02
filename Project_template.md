## Изучите [README.md](README.md) файл и структуру проекта.

## Задание 1

[Диаграмма контейнеров C4 (PlantUML)](./architecture/to-be-c4-container.puml)
![Диаграмма контейнеров C4](./architecture/to-be-c4-container.png)


## Задание 2

### 1. Proxy
Реализован прокси
Протестирован постепенный переход 
MOVIES_MIGRATION_PERCENT=50%
```bash
➜  ya-practicum-02 docker attach cinemaabyss-proxy-service
2026/02/28 12:52:06 [monolith] GET /health -> http://monolith:8080
2026/02/28 12:52:06 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:52:06 [monolith] GET /api/movies -> http://monolith:8080
2026/02/28 12:52:07 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:52:07 [monolith] GET /api/movies -> http://monolith:8080
2026/02/28 12:52:08 [monolith] GET /api/movies -> http://monolith:8080
2026/02/28 12:52:09 [monolith] GET /api/movies -> http://monolith:8080
2026/02/28 12:52:09 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:52:10 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:52:10 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:52:11 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:52:11 [monolith] GET /api/users -> http://monolith:8080
2026/02/28 12:52:11 [monolith] GET /api/payments -> http://monolith:8080
```

MOVIES_MIGRATION_PERCENT=100%
```bash
➜  ya-practicum-02 git:(cinema) ✗ docker logs cinemaabyss-proxy-service
2026/02/28 12:55:59 Starting Proxy Service (API Gateway) on port 8000
2026/02/28 12:55:59 Monolith URL: http://monolith:8080
2026/02/28 12:55:59 Movies Service URL: http://movies-service:8081
2026/02/28 12:55:59 Events Service URL: http://events-service:8082
2026/02/28 12:55:59 Gradual Migration: true
2026/02/28 12:55:59 Movies Migration Percent: 100%
2026/02/28 12:56:07 Starting Proxy Service (API Gateway) on port 8000
2026/02/28 12:56:07 Monolith URL: http://monolith:8080
2026/02/28 12:56:07 Movies Service URL: http://movies-service:8081
2026/02/28 12:56:07 Events Service URL: http://events-service:8082
2026/02/28 12:56:07 Gradual Migration: true
2026/02/28 12:56:07 Movies Migration Percent: 100%
2026/02/28 12:56:42 [monolith] GET /health -> http://monolith:8080
2026/02/28 12:56:42 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:56:43 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:56:43 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:56:44 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:56:44 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:56:45 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:56:46 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:56:46 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:56:47 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:56:47 [movies-service] GET /api/movies -> http://movies-service:8081
2026/02/28 12:56:48 [monolith] GET /api/users -> http://monolith:8080
2026/02/28 12:56:48 [monolith] GET /api/payments -> http://monolith:8080
```

### 2. Kafka

Разработан сервис
Тесты прошли 
![Результат](./events-service-test-result.png)
![Результат](./event-service-test-kafka.png)

## Задание 3

Результат вызоыв api/movies
![Результат вызоыв api/movies](./movies-response.png)

Логи events-servise
![Логи events-servise](./events-service-kuber-logs.png)

## Задание 4
Для простоты дальнейшего обновления и развертывания вам как архитектуру необходимо так же реализовать helm-чарты для прокси-сервиса и проверить работу 

Для этого:
1. Перейдите в директорию helm и отредактируйте файл values.yaml

```yaml
# Proxy service configuration
proxyService:
  enabled: true
  image:
    repository: ghcr.io/db-exp/cinemaabysstest/proxy-service
    tag: latest
    pullPolicy: Always
  replicas: 1
  resources:
    limits:
      cpu: 300m
      memory: 256Mi
    requests:
      cpu: 100m
      memory: 128Mi
  service:
    port: 80
    targetPort: 8000
    type: ClusterIP
```

- Вместо ghcr.io/db-exp/cinemaabysstest/proxy-service напишите свой путь до образа для всех сервисов
- для imagePullSecret проставьте свое значение (скопируйте из конфигурации kubernetes)
  ```yaml
  imagePullSecrets:
      dockerconfigjson: ewoJImF1dGhzIjogewoJCSJnaGNyLmlvIjogewoJCQkiYXV0aCI6ICJaR0l0Wlhod09tZG9jRjl2UTJocVZIa3dhMWhKVDIxWmFVZHJOV2hRUW10aFVXbFZSbTVaTjJRMFNYUjRZMWM9IgoJCX0KCX0sCgkiY3JlZHNTdG9yZSI6ICJkZXNrdG9wIiwKCSJjdXJyZW50Q29udGV4dCI6ICJkZXNrdG9wLWxpbnV4IiwKCSJwbHVnaW5zIjogewoJCSIteC1jbGktaGludHMiOiB7CgkJCSJlbmFibGVkIjogInRydWUiCgkJfQoJfSwKCSJmZWF0dXJlcyI6IHsKCQkiaG9va3MiOiAidHJ1ZSIKCX0KfQ==
  ```

2. В папке ./templates/services заполните шаблоны для proxy-service.yaml и events-service.yaml (опирайтесь на свою kubernetes конфигурацию - смысл helm'а сделать шаблоны для быстрого обновления и установки)

```yaml
template:
    metadata:
      labels:
        app: proxy-service
    spec:
      containers:
       Тут ваша конфигурация
```

3. Проверьте установку
Сначала удалим установку руками

```bash
kubectl delete all --all -n cinemaabyss
kubectl delete  namespace cinemaabyss
```
Запустите 
```bash
helm install cinemaabyss .\src\kubernetes\helm --namespace cinemaabyss --create-namespace
```
Если в процессе будет ошибка
```code
[2025-04-08 21:43:38,780] ERROR Fatal error during KafkaServer startup. Prepare to shutdown (kafka.server.KafkaServer)
kafka.common.InconsistentClusterIdException: The Cluster ID OkOjGPrdRimp8nkFohYkCw doesn't match stored clusterId Some(sbkcoiSiQV2h_mQpwy05zQ) in meta.properties. The broker is trying to join the wrong cluster. Configured zookeeper.connect may be wrong.
```

Проверьте развертывание:
```bash
kubectl get pods -n cinemaabyss
minikube tunnel
```

Потом вызовите 
https://cinemaabyss.example.com/api/movies
и приложите скриншот развертывания helm и вывода https://cinemaabyss.example.com/api/movies


# Задание 5
Компания планирует активно развиваться и для повышения надежности, безопасности, реализации сетевых паттернов типа Circuit Breaker и канареечного деплоя вам как архитектору необходимо развернуть istio и настроить circuit breaker для monolith и movies сервисов.

```bash

helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo update

helm install istio-base istio/base -n istio-system --set defaultRevision=default --create-namespace
helm install istio-ingressgateway istio/gateway -n istio-system
helm install istiod istio/istiod -n istio-system --wait

helm install cinemaabyss .\src\kubernetes\helm --namespace cinemaabyss --create-namespace

kubectl label namespace cinemaabyss istio-injection=enabled --overwrite

kubectl get namespace -L istio-injection

kubectl apply -f .\src\kubernetes\circuit-breaker-config.yaml -n cinemaabyss

```

Тестирование

# fortio
```bash
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.25/samples/httpbin/sample-client/fortio-deploy.yaml -n cinemaabyss
```

# Get the fortio pod name
```bash
FORTIO_POD=$(kubectl get pod -n cinemaabyss | grep fortio | awk '{print $1}')

kubectl exec -n cinemaabyss $FORTIO_POD -c fortio -- fortio load -c 50 -qps 0 -n 500 -loglevel Warning http://movies-service:8081/api/movies
```
Например,

```bash
kubectl exec -n cinemaabyss fortio-deploy-b6757cbbb-7c9qg  -c fortio -- fortio load -c 50 -qps 0 -n 500 -loglevel Warning http://movies-service:8081/api/movies
```

Вывод будет типа такого

```bash
IP addresses distribution:
10.106.113.46:8081: 421
Code 200 : 79 (15.8 %)
Code 500 : 22 (4.4 %)
Code 503 : 399 (79.8 %)
```
Можно еще проверить статистику

```bash
kubectl exec -n cinemaabyss fortio-deploy-b6757cbbb-7c9qg -c istio-proxy -- pilot-agent request GET stats | grep movies-service | grep pending
```

И там смотрим 

```bash
cluster.outbound|8081||movies-service.cinemaabyss.svc.cluster.local;.upstream_rq_pending_total: 311 - столько раз срабатывал circuit breaker
You can see 21 for the upstream_rq_pending_overflow value which means 21 calls so far have been flagged for circuit breaking.
```

Приложите скриншот работы circuit breaker'а

Удаляем все
```bash
istioctl uninstall --purge
kubectl delete namespace istio-system
kubectl delete all --all -n cinemaabyss
kubectl delete namespace cinemaabyss
```
