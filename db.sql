-- Включить расширение pgcrypto для генерации UUID (если еще не включено)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);
-- Таблица банковских счетов
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    balance NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- Таблица транзакций
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    operation TEXT NOT NULL, -- "deposit", "withdraw", "transfer_in", "transfer_out"
    amount NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);
-- Таблица банковских карт
CREATE TABLE cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    number_enc TEXT NOT NULL,         -- PGP-шифрованный номер
    expiry_month_enc TEXT NOT NULL,   -- PGP-шифрованный месяц
    expiry_year_enc TEXT NOT NULL,    -- PGP-шифрованный год
    cvv_hash TEXT NOT NULL,           -- bcrypt от CVV
    hmac TEXT NOT NULL,               -- HMAC всех полей
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Таблица кредитов
CREATE TABLE credits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    amount NUMERIC(12,2) NOT NULL,
    term_months INT NOT NULL,
    interest_rate NUMERIC(5,2) NOT NULL, -- Годовая ставка (например, 12.5)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Таблица графика платежей
CREATE TABLE payment_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_id UUID NOT NULL REFERENCES credits(id) ON DELETE CASCADE,
    due_date DATE NOT NULL,
    amount NUMERIC(12,2) NOT NULL,
    paid BOOLEAN DEFAULT FALSE,
    paid_at TIMESTAMP
);