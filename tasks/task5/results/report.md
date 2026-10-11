# Задание 5. Отчёт

## Что сделано

Скопировал задание 4 в `task5` и развернул `booking-service` в Minikube с Istio (профиль `demo`). В namespace `default` включил инъекцию sidecar, поэтому каждый под работает с прокси Envoy.

### Две версии сервиса

| | v1 (`values-v1.yaml`) | v2 (`values-v2.yaml`) |
|---|---|---|
| Реплик | 2 | 1 |
| `ENABLE_FEATURE_X` | false | true, работает `GET /feature` |
| Метки пода | `app=booking-service`, `version=v1` | `app=booking-service`, `version=v2` |

Обе версии разворачиваются одним Helm-чартом. Общий Service `booking-service` выбирает поды обеих версий по метке `app`. Чтобы версии можно было различить, сервис возвращает свою версию в ответе `GET /version` и в заголовке `X-App-Version`.

### Маршрутизация (`istio/`)

- **Canary:** `VirtualService` отправляет 90% запросов на v1 и 10% на v2.
- **Retry:** `VirtualService` повторяет запрос до 3 раз при ошибке.
- **Circuit breaking:** `DestinationRule` ограничивает число соединений и на 30 секунд исключает под после трёх ошибок 5xx подряд.
- **Fallback:** `DestinationRule` даёт подам v1 приоритет (`failoverPriority`), поэтому запросы для v1 уходят на v2, только когда исправных подов v1 не осталось.
- **Фича-флаг:** `EnvoyFilter` направляет на v2 запросы с заголовком `X-Feature-Enabled: true`.

Пайплайн из задания 4 разворачивает обе версии и применяет манифесты из `istio/`.

## Запуск

Из `tasks/task5` с запущенным Minikube:

```bash
istioctl install --set profile=demo -y
kubectl label namespace default istio-injection=enabled
gitlab-ci-local build deploy
kubectl -n istio-system port-forward svc/istio-ingressgateway 9090:80
```

Проверочные скрипты обращаются к ingress gateway через `localhost:9090`.

## Проверки

| Скрипт | Результат |
|---|---|
| `check-istio.sh` | Istio работает, инъекция в `default` включена |
| `check-canary.sh` | Из 100 запросов 91 обработала v1, 9 - v2 |
| `check-feature-flag.sh` | С заголовком все запросы уходят на v2, без заголовка `/feature` возвращает 404 |
| `check-fallback.sh` | Без одного пода v1 все запросы успешны, без всех подов v1 все запросы обслуживает v2 |

## Файлы результата

| Файл | Содержание |
|---|---|
| [values-v1.yaml](values-v1.yaml), [values-v2.yaml](values-v2.yaml) | Конфигурации версий |
| [virtual-service.yaml](virtual-service.yaml) | Canary, fallback, фича-флаг, retry |
| [destination-rule.yaml](destination-rule.yaml) | Subsets, circuit breaking, приоритет v1 |
| [envoy-filter.yaml](envoy-filter.yaml) | Фича-флаг через EnvoyFilter |
| [check-istio.txt](check-istio.txt), [check-canary.txt](check-canary.txt), [check-feature-flag.txt](check-feature-flag.txt), [check-fallback.txt](check-fallback.txt) | Логи проверочных скриптов |
| [pipeline-log.txt](pipeline-log.txt) | Лог пайплайна |
