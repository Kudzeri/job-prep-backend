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

Для pet-проекта есть `render.yaml`: подключите репозиторий в Render через **New → Blueprint**, выберите бесплатные планы и укажите `INTEGRATION_API_KEY` при запросе. Render соберет Dockerfile, создаст PostgreSQL и развернет backend `job-prep-api` и статическую панель `prepwise-dashboard`. Если Render назначил панели другой домен, измените `CORS_ORIGINS` в настройках backend на её точный origin, например `https://prepwise-dashboard.onrender.com` (без `/` в конце).

Открой панель по URL статического сервиса и зайди в **Подключение API**. В поле Backend URL введи URL backend-сервиса Render и нажми **Сохранить и проверить**. Затем войди в аккаунт, чтобы видеть личные тесты; API-ключ в панели для этого не требуется. Если нужно просматривать или создавать тесты общей интеграционной библиотеки, дополнительно введи в `X-API-Key` значение `INTEGRATION_API_KEY`. Для запросов со статической панели backend должен разрешать её origin через `CORS_ORIGINS` и заголовок `X-API-Key`.

После входа по email или Telegram панель показывает личные тесты пользователя. Они защищены JWT и не видны другим аккаунтам; тесты, созданные внешним API-ключом, остаются в общей интеграционной библиотеке. В режиме личного аккаунта API-ключ в панели не нужен.

Список и содержимое тестов доступны на API с API-ключом через `GET /api/v1/integrations/tests` и `GET /api/v1/integrations/tests/{testId}`. В панели можно просматривать тесты, создавать тест и отправлять пакет вопросов. API-ключ, введённый в браузере, доступен пользователю браузера; это прототип интерфейса для ручной проверки, а не схема защиты для публичного пользовательского продукта. Вход по email-коду требует настройки `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD` и `SMTP_FROM`; Telegram-вход требует `TELEGRAM_BOT_TOKEN` и `TELEGRAM_BOT_USERNAME`.

Swagger доступен по адресу `https://<backend-сервис>.onrender.com/swagger`. В Swagger нажми **Authorize** и введи API-ключ.

## Чат и внешний сервис генерации тестов

В интерфейсе AI-чат отвечает готовыми фразами на приветствие, вопрос «как дела», благодарность и вопросы о возможностях. Для прочих тем он честно сообщает, что пока ограничен. Запрос на персональный тест требует входа, резюме PDF и допускает ещё два материала PDF/DOCX/TXT/MD. Максимум 10 МБ на файл и 24 МБ суммарно.

Генератор работает через внешний асинхронный сервис, которого в этом репозитории нет. Пока его адрес не задан, API вернёт `503` и чат сообщит, что генерация не подключена. В настройках backend на Render добавь:

- `PUBLIC_API_URL` — публичный корень этого backend, например `https://job-prep-api.onrender.com`.
- `ASSESSMENT_SERVICE_URL` — корень внешнего сервиса генерации, без `/requests` в конце.
- `ASSESSMENT_SERVICE_API_KEY` — необязательный токен для заголовка `Authorization: Bearer ...` внешнего сервиса.
- `ASSESSMENT_CALLBACK_SECRET` — случайный общий секрет для callback.

После изменений перезапусти backend. При запросе чата backend отправит на `POST {ASSESSMENT_SERVICE_URL}/requests` multipart-поля `request_id`, `callback_url`, `message`, файл `resume` и до двух повторяющихся файлов `attachments`. Запрос аутентифицируется опциональным Bearer-токеном и заголовком `X-Assessment-Callback-Secret`. Внешний сервис должен быстро принять запрос с любым `2xx`, обработать файлы и отправить результат на `callback_url`:

```http
POST /api/v1/integrations/assistant/callback
X-Assessment-Callback-Secret: <ASSESSMENT_CALLBACK_SECRET>
Content-Type: application/json
```

Ответ текстом:

```json
{"request_id":"88e01249-2713-4bca-b0cc-72222483e8af","result":{"type":"message","message":"Добавь описание вакансии, чтобы тест был точнее."}}
```

Ответ с готовым тестом (вопросы должны содержать `text`; `options` и `answer` передаются как JSON):

```json
{"request_id":"88e01249-2713-4bca-b0cc-72222483e8af","result":{"type":"test","title":"Подготовка к интервью","description":"Вопросы по резюме","questions":[{"text":"Чем отличается map от sync.Map?","options":["map не потокобезопасен","оба потокобезопасны"],"answer":"map не потокобезопасен"}]}}
```

Ошибка обработки: `{"request_id":"…","status":"failed","result":{"type":"message","message":"Не удалось прочитать резюме"}}`. Результат test сохраняется в личные тесты пользователя атомарно; чат проверяет его через `GET /api/v1/assistant/requests/{requestId}`. Полные тела, ограничения и схемы доступны в Swagger.

На бесплатном тарифе веб-сервис засыпает после 15 минут без трафика, поэтому первый запрос может запускаться около минуты. Бесплатная база Render действует 30 дней и ограничена 1 ГБ; этот вариант подходит для временного теста, не для хранения важных данных. [Ограничения бесплатного тарифа Render](https://render.com/docs/free).
