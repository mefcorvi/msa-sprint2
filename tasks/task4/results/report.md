# Задание 4. Отчёт

## Что сделано

### Сервис

Заменил заглушку на `booking-service` из задания 2. Для Kubernetes добавил в сервис:

- HTTP-эндпоинты на порту 8080: `/ping` отвечает `pong` и используется в пробах, а `/health` и `/ready` показывают, работает ли процесс и доступна ли БД;
- фича-флаг `ENABLE_FEATURE_X`: при `true` работает `GET /feature`, иначе этот запрос получает 404;
- запуск без БД: сервис стартует сразу и подключается к PostgreSQL в фоне, поэтому под не падает, пока БД недоступна.

В Minikube развернул только сервис, так как `/ping` и проверка DNS от БД не зависят.

### Helm-чарт (`helm/booking-service`)

- Deployment проверяет `/ping` в `livenessProbe` и `readinessProbe`, а переменные окружения и ресурсы берёт из values.
- Service типа `ClusterIP` открывает порт 80 для HTTP и порт 9090 для gRPC.
- `values-staging.yaml` задаёт 1 реплику и включает флаг, `values-prod.yaml` задаёт 3 реплики и выключает флаг.

### CI/CD (`.gitlab-ci.yml`)

| Стадия | Job | Что делает |
|---|---|---|
| build | `build` | Собирает образ с тегом по SHA коммита |
| test | `unit-tests` | Запускает unit-тесты Go |
| test | `test` | Запускает контейнер и проверяет, что `/ping` отвечает `pong` |
| deploy | `deploy` | Загружает образ в Minikube и выполняет `helm upgrade --install` |
| tag | `tag` | Создаёт git-тег `booking-service-<время>` |

Пайплайн запускается командой `gitlab-ci-local`. Файл значений задаёт переменная `VALUES_FILE`, по умолчанию это `values-staging.yaml`.

### Service Discovery

Внутри кластера сервис доступен по DNS-имени `booking-service`: `check-dns.sh` из отдельного пода получает `pong` по адресу `http://booking-service/ping`.

## Файлы результата

| Файл | Содержание |
|---|---|
| [values-staging.yaml](values-staging.yaml), [values-prod.yaml](values-prod.yaml) | Значения чарта для staging и prod |
| [.gitlab-ci.yml](.gitlab-ci.yml) | Пайплайн |
| [pipeline-log.txt](pipeline-log.txt) | Лог успешного запуска пайплайна |
| [curl-ping.png](curl-ping.png) | Успешный `curl` на `/ping` через `kubectl port-forward` |
| [check-dns.png](check-dns.png) | Вывод `check-dns.sh`: под в кластере получает `pong` по DNS-имени `booking-service` |
| [check-status.png](check-status.png) | Вывод `check-status.sh`: под, Service, Helm-релиз и ответ `/ping` |
| [kubectl-get.txt](kubectl-get.txt) | `kubectl get pods,services` |
| [images.txt](images.txt) | `docker image ls` и `minikube image list` |

## Замечания

- Инструкция по развёртыванию находится в `tasks/task4/README.md`.
