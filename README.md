

# Структура проекта
```
fin_service/
├── main.go                         # Точка входа: запуск сервера, инициализация компонентов
├── go.mod                          # Go-модули и зависимости
├── .env                            # Конфигурация: SMTP, DB, JWT_SECRET и т.д.

├── integration/
│   └── cbr/
│       ├── cbr.go                  # Модуль получения ставки ЦБ
│       └── cbr_test.go

├── internal/
│   ├── handler/
│   │   ├── register_handler.go     # POST /register
│   │   ├── login_handler.go        # POST /login
│   │   ├── account_handler.go      # POST /accounts, /{id}/deposit, /{id}/withdraw, GET /accounts, /balance
│   │   └── transfer_handler.go     # POST /transfers

│   ├── service/
│   │   ├── user_service.go         # Регистрация: проверка, хеш, welcome-письмо
│   │   ├── auth_service.go         # Аутентификация: JWT, письмо о входе
│   │   ├── account_service.go      # Счета: логика создания, операций, писем, логов
│   │   └── transfer_service.go     # Переводы: валидация, письма, вызов транзакции

│   ├── repository/
│   │   ├── user_repo.go            # Пользователи: поиск, проверка, email
│   │   ├── account_repo.go         # Счета: работа с балансами
│   │   ├── transaction_repo.go     # INSERT в transactions
│   │   └── transfer_repo.go        # Перевод: транзакция изменения балансов + логи

│   ├── integration/
│   │   └── smtp.go                 # Отправка писем: welcome, вход, депозит, снятие, переводы

│   └── models/
│       ├── user.go                 # Структура User, JSON, валидация
│       ├── account.go              # Структура Account
│       └── transaction.go          # Структура Transaction

├── pkg/                            # Общие библиотеки (если появятся)
│
└── utils/
    └── env.go                      # Загрузка переменных окружения из .env


```

## Регистрация пользователя и отправка письма

Реализован полноценный процесс регистрации пользователя с отправкой приветственного письма.

### POST /register

**JSON-запрос:**
```json
{
  "email": "user@example.com",
  "username": "yourname",
  "password": "StrongPassword123"
}
```
Функциональность:

- Проверка формата email, длины пароля (8+ символов), уникальности email/username.
- Хеширование пароля с использованием bcrypt.
- Сохранение пользователя в базу данных PostgreSQL (users таблица, UUID).
- Отправка приветственного письма на email через SMTP.

SMTP

- Используется gomail.v2 для отправки писем.
- Конфигурация через .env:
```bash
SMTP_HOST=smtp.yandex.ru
SMTP_PORT=465
SMTP_USER=your@yandex.ru
SMTP_PASSWORD=your_password
SMTP_FROM=your@yandex.ru
```
Ответ при успешной регистрации:
```json
{
  "message": "Registration successful, welcome email sent."
}
```
При ошибке валидации или дублировании email/username возвращается 400 Bad Request с описанием ошибки.

## Аутентификация пользователя и отправка письма

Реализован вход пользователя через endpoint `/login` с генерацией JWT и отправкой письма при успешной авторизации.

### POST /login

**JSON-запрос:**
```json
{
  "email": "user@example.com",
  "password": "StrongPassword123"
}
```
Функциональность:

- Поиск пользователя по email в PostgreSQL
- Сравнение пароля с хешем (bcrypt)
- Генерация JWT-токена на 24 часа (алгоритм HS256)
- Отправка письма "Вы успешно вошли в систему" через SMTP

Ответ:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```
Настройки

JWT секрет берётся из .env:

JWT_SECRET=super-strong-random-secret-key

SMTP отправка работает так же, как при регистрации:

SMTP_FROM=your@yandex.ru

Ошибки:

- 401 Unauthorized: неправильные email или пароль
- 500 Internal Server Error: ошибки внутри сервиса, проблемы с SMTP или JWT

## Создание счёта пользователя

Реализован защищённый endpoint `/accounts` для создания банковского счёта.

### POST /accounts

**Требуется аутентификация (JWT)**

**Заголовок:**
```http
Authorization: Bearer <токен_из_login>
```

**Тело запроса:**
```json
{}
```
(Можно оставить пустым — счёт создаётся в RUB с нулевым балансом.)

Ответ:
```json
{
  "account_id": "3e748fc0-6a89-4d12-a950-2c32cd66101c"
}
```
Поведение:

- Только авторизованный пользователь может создать счёт
- Валюта по умолчанию: RUB
- Начальный баланс: 0.00
- Один пользователь может иметь несколько счетов
- Все данные сохраняются в таблицу accounts (PostgreSQL)
- Пользователю отправляется email с темой "New Bank Account Created"
- Email берётся по user_id через таблицу users

Таблица accounts
```asql
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    balance NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```
Ошибки:
- `401 Unauthorized` — если отсутствует или неверен JWT
- `500 Internal Server Error` — шибка при создании счёта или отправке email

## Просмотр счетов и баланса

Реализованы два защищённых эндпоинта для получения информации о банковских счетах пользователя.

### GET /accounts

Возвращает список всех банковских счетов текущего пользователя.

**Заголовок:**
```http
Authorization: Bearer <jwt_token>
```

**Ответ:**
```json
[
  {
    "account_id": "3e748fc0-6a89-4d12-a950-2c32cd66101c",
    "balance": 1000.50
  },
  ...
]
```
```http
GET /accounts/{id}/balance
```
Возвращает баланс конкретного счёта (только если он принадлежит текущему пользователю).

Заголовок:
```http
Authorization: Bearer <jwt_token>
```
Пример URL:
```http
GET /accounts/3e748fc0-6a89-4d12-a950-2c32cd66101c/balance
```
Ответ:
```json
{
  "account_id": "3e748fc0-6a89-4d12-a950-2c32cd66101c",
  "balance": 1000.50
}
```
Ошибки:

- `401 Unauthorized` — если токен не указан или недействителен
- `403 Forbidden` — если пользователь запрашивает чужой счёт
- `404 Not Found` — если счёт не существует

## Пополнение счёта

Реализован endpoint `POST /accounts/{id}/deposit` для пополнения счёта текущего пользователя.

### POST /accounts/{id}/deposit

**Требуется аутентификация (JWT)**

**Пример запроса:**
```http
POST /accounts/3e748fc0-6a89-4d12-a950-2c32cd66101c/deposit
Authorization: Bearer <jwt_token>
```
Тело запроса:
```json
{
  "amount": 1000.00
}
```
Ответ:
```json
{
  "message": "Deposit successful",
  "new_balance": 1500.00
}
```
Поведение:
- Сумма должна быть положительной (> 0)
- Пользователь может пополнять только свои счета
- Баланс увеличивается на указанную сумму
- Пользователю отправляется письмо на email с темой: "Bank account deposit"

Ошибки:
- `401 Unauthorized` — отсутствует или неверный токен
- `403 Forbidden` — попытка пополнить чужой счёт
- `400 Bad Request` — неверный JSON или сумма <= 0

## Снятие средств со счёта

Реализован endpoint `POST /accounts/{id}/withdraw` для снятия средств со счёта текущего пользователя.

### POST /accounts/{id}/withdraw

**Требуется аутентификация (JWT)**

**Пример запроса:**
```http
POST /accounts/3e748fc0-6a89-4d12-a950-2c32cd66101c/withdraw
Authorization: Bearer <jwt_token>
```
Тело запроса:
```json
{
  "amount": 500.00
}
```
Ответ:
```json
{
  "message": "Withdrawal successful",
  "new_balance": 1000.00
}
```
Поведение:
- Сумма должна быть положительной (> 0)
- Пользователь может снимать средства только со своих счетов
- Проверяется наличие достаточного баланса
- Баланс уменьшается на указанную сумму
- Пользователю отправляется письмо на email с темой: "Bank account withdrawal"

Ошибки:
- `401 Unauthorized` — отсутствует или неверный токен
- `403 Forbidden` — попытка снять средства с чужого счёта
- `400 Bad Request` — неверный JSON или сумма <= 0
- `400 Bad Request` — недостаточно средств на счёте

## Перевод между счетами

Реализован защищённый endpoint `POST /transfers` для перевода средств между двумя банковскими счетами.

### POST /transfers

**Требуется аутентификация (JWT)**

**Пример запроса:**
```json
{
  "from_account_id": "5c244e21-d1ad-45b6-bc3a-6789894e3222",
  "to_account_id":   "bd9cfef2-d3c5-4081-b420-c4c383aa1004",
  "amount": 500.00
}
```

Пример ответа:
```json
{
  "message": "Transfer successful"
}
```
Условия и проверки:

- Отправитель может переводить только со своего счёта
- Нельзя переводить на тот же счёт
- Сумма должна быть положительной (> 0)
- Баланс отправителя должен быть достаточным
- Перевод выполняется в одной транзакции
- В transactions создаются записи:
    - transfer_out — со счёта отправителя
    - transfer_in — на счёт получателя
- Оба пользователя получают email-уведомления:
    - Отправителю: "Bank account transfer sent"
    - Получателю: "Bank account transfer received"

Ошибки:
- `401 Unauthorized` — отсутствует токен
- `403 Forbidden` — попытка перевода с чужого счёта
- `400 Bad Request` — недостаточно средств, неверный формат, одинаковые счета

## История операций по счёту

Реализован endpoint `GET /accounts/{id}/transactions`, возвращающий все операции по счёту в хронологическом порядке (от новых к старым).

### GET /accounts/{id}/transactions

**Требуется аутентификация (JWT)**

**Пример запроса:**
```http
GET /accounts/3e748fc0-6a89-4d12-a950-2c32cd66101c/transactions
Authorization: Bearer <jwt_token>
```

**Пример ответа:**
```json
[
  {
    "id": "f91c2a47-d50e-40f7-8c9e-bbd9d0a0c177",
    "account_id": "3e748fc0-6a89-4d12-a950-2c32cd66101c",
    "operation": "deposit",
    "amount": 1000.00,
    "description": "User-initiated deposit",
    "created_at": "2025-05-02T13:45:12Z"
  },
  {
    "id": "...",
    "operation": "transfer_out",
    "amount": 500.00,
    ...
  }
]
```

Возможные операции:
- "deposit" — пополнение счёта
- "withdraw" — снятие
- "transfer_in" — входящий перевод
- "transfer_out" — исходящий перевод

Ошибки:
- `401 Unauthorized` — отсутствует токен
- `403 Forbidden` — попытка доступа к чужому счёту
- `404 Not Found` — счёт не найден

## Аналитика по счёту

Реализован endpoint `GET /accounts/{id}/analytics`, возвращающий сумму доходов и расходов за указанный месяц.

### GET /accounts/{id}/analytics?month=YYYY-MM

**Требуется аутентификация (JWT)**

**Пример запроса:**
```http
GET /accounts/414ac401-f3c7-42f5-b500-ae10eaca4d2a/analytics?month=2025-05
Authorization: Bearer <jwt_token>
```

**Пример ответа:**
```json
{
  "account_id": "414ac401-f3c7-42f5-b500-ae10eaca4d2a",
  "month": "2025-05",
  "income": 1500.00,
  "expense": 700.00
}
```
Правила расчёта:

- Доходы: суммы по операциям deposit, transfer_in
- Расходы: суммы по withdraw, transfer_out
- Выводятся операции только за указанный месяц (YYYY-MM)
- Пользователь может запрашивать только свои счета

Ошибки:
- `400 Bad Request` — если отсутствует параметр month, неверный формат (YYYY-MM) или счёт не принадлежит пользователю
- `401 Unauthorized` — если отсутствует токен

## Генерация виртуальной карты

Реализован endpoint `POST /cards`, создающий виртуальную карту, привязанную к банковскому счёту пользователя.

### POST /cards

**Требуется аутентификация (JWT)**

**Пример запроса:**
```json
{
  "account_id": "414ac401-f3c7-42f5-b500-ae10eaca4d2a"
}
```

Пример ответа:
```json
{
  "card_id": "d5a94cb4-d0ad-4b39-bfc3-6fbe5d96eb71",
  "created_at": "2025-05-02 18:11:45",
  "expiry_month": 5,
  "expiry_year": 2028
}
```
Что происходит:

- Генерируется валидный номер карты (алгоритм Луна)
- Срок действия: 3 года вперёд от текущей даты
- CVV: случайный 3-значный код
- Данные защищены:
  - Номер, срок действия — PGP (временно эмуляция)
  - CVV — bcrypt
  - HMAC — контроль целостности

Карта сохраняется в таблицу cards

Ошибки:
- `400 Bad Request` — если account_id не передан или недействителен
- `403 Forbidden` — если счёт не принадлежит пользователю
- `401 Unauthorized` — при отсутствии или ошибке в JWT

## Просмотр виртуальных карт

Реализован endpoint `GET /cards`, возвращающий все карты, выпущенные текущим пользователем.

### GET /cards

**Требуется аутентификация (JWT)**

**Пример запроса:**
```http
GET /cards
Authorization: Bearer <jwt_token>
```

**Пример ответа:**
```json
{
  "cards": [
    {
      "card_id": "87bb4e19-63ba-4bbc-abe6-3356dc6b1a08",
      "account_id": "414ac401-f3c7-42f5-b500-ae10eaca4d2a",
      "created_at": "2025-05-03 11:14:54",
      "masked_number": "**** **** **** 9614",
      "expiry": "05/28"
    }
  ]
}
```
Что отображается:

- Последние 4 цифры карты
- Срок действия (расшифровка из зашифрованных данных)
- ID карты и привязанного счёта
- Дата выпуска
CVV и полный номер карты не возвращаются.

Ошибки:
- `401 Unauthorized` — если отсутствует или невалиден токен
- `500 Internal Server Error` — при ошибке чтения из БД


## Оплата с карты

Реализован endpoint `POST /cards/{card_id}/pay`, позволяющий списать средства с карты (привязанного счёта).

### POST /cards/{card_id}/pay

**Пример запроса:**
```json
{
  "amount": 250.00,
  "description": "Payment for order #1234"
}
```
Пример ответа:
```json
{
  "message": "Payment successful"
}
```
Поведение:

- Проверяется срок действия карты
- Списание происходит с привязанного счёта
- Операция логируется в transactions как withdraw
- Пользователю отправляется email об оплате

Ошибки:
- `401 Unauthorized` — при отсутствии токена
- `403 Forbidden` — если карта чужая
- `400 Bad Request` — если карта истекла или средств недостаточно