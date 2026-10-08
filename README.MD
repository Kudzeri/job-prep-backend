# Job Prep Backend

Swagger UI доступен после запуска приложения по адресу:

- `/swagger`
- `/swagger/openapi.yaml`

Для защищенных эндпоинтов используется заголовок `Authorization: Bearer <access_token>`.

## Внешнее создание тестов

Для вызовов с другого сайта задайте секрет `INTEGRATION_API_KEY` в окружении API. Передавайте его сервер-сервер заголовком `X-API-Key`; не встраивайте этот секрет в публичный JavaScript-код браузера. Если интеграция запускается прямо из браузера, добавьте его origin в `CORS_ORIGINS`.

`POST /api/v1/integrations/tests` создает тест. Ответ содержит `id`; его передают для добавления вопросов:

```http
POST /api/v1/integrations/tests/42/questions
X-API-Key: <INTEGRATION_API_KEY>
Content-Type: application/json

{"text":"Что такое горутина?","options":["поток выполнения","тип данных"],"answer":"поток выполнения"}
```

Для пакетной вставки используется `POST /api/v1/integrations/tests/42/questions/batch` с телом `{"questions":[...]}`. Пакет может содержать до 500 вопросов и записывается атомарно. Полная схема и примеры доступны в Swagger.

## Kubernetes и CI/CD

Образ собирается Dockerfile. Workflow `.github/workflows/deploy.yml` при push в `main` публикует образ в GHCR и разворачивает `k8s/` в кластер. В GitHub Actions необходимо добавить secret `KUBE_CONFIG` (base64 kubeconfig). До первого запуска создайте секрет приложения и обеспечьте доступ к PostgreSQL:

```sh
kubectl apply -f k8s/namespace.yaml
kubectl -n job-prep create secret generic job-prep-api-secrets \
  --from-literal=DB_URL='postgres://USER:PASSWORD@HOST:5432/DB?sslmode=require' \
  --from-literal=JWT_SECRET='replace-with-a-long-random-secret' \
  --from-literal=INTEGRATION_API_KEY='replace-with-a-long-random-secret' \
  --from-literal=SMTP_USERNAME='' --from-literal=SMTP_PASSWORD='' \
  --from-literal=TELEGRAM_BOT_TOKEN=''
```

Замените `CORS_ORIGINS` в `k8s/configmap.yaml` на origin вашего сайта. PostgreSQL в Kubernetes этим набором манифестов не разворачивается: `DB_URL` должен указывать на доступную кластеру базу.

`k8s/ingress.yaml` рассчитан на ingress-nginx и TLS secret `job-prep-api-tls`: замените `api.your-domain.example` и создайте TLS secret до деплоя. Для приватного пакета GHCR настройте `imagePullSecret` для namespace либо сделайте пакет доступным кластеру.

## Быстрый бесплатный тестовый деплой

Для pet-проекта есть `render.yaml`: подключите репозиторий в Render через **New → Blueprint**, выберите бесплатные планы и укажите `INTEGRATION_API_KEY` при запросе. Render соберет Dockerfile, создаст PostgreSQL и выдаст публичный HTTPS URL. После успешного деплоя откройте `https://<имя-сервиса>.onrender.com/swagger`, в Swagger нажмите **Authorize** и введите API-ключ; друг сможет отправлять запросы прямо из документации.

На бесплатном тарифе веб-сервис засыпает после 15 минут без трафика, поэтому первый запрос может запускаться около минуты. Бесплатная база Render действует 30 дней и ограничена 1 ГБ; этот вариант подходит для временного теста, не для хранения важных данных. [Ограничения бесплатного тарифа Render](https://render.com/docs/free).
