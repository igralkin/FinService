package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"fin_service/internal/handler"
	"fin_service/internal/integration"
	"fin_service/internal/integration/cbr"
	"fin_service/internal/repository"
	"fin_service/internal/service"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found or it could not be loaded")
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func initDB() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	return sql.Open("postgres", dsn)
}

func main() {
	// --- Получение ставки ЦБ (текущий функционал) ---
	rate, err := cbr.GetCentralBankRate()
	if err != nil {
		log.Fatalf("Failed to get key rate: %v", err)
	}
	fmt.Printf("Current Central Bank key rate: %.2f%%\n", rate)

	// --- Инициализация SMTP ---
	smtpClient, err := integration.NewSMTPClientFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize SMTP client: %v", err)
	}

	// --- Подключение к базе данных PostgreSQL ---
	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// --- Хранилище пользователей (PostgreSQL) ---
	userRepo := repository.NewUserRepository(db)

	// --- Сервисы ---
	userService := service.NewUserService(userRepo, smtpClient)
	authService := service.NewAuthService(userRepo, smtpClient)

	// --- Обработчики ---
	// Публичные
	registerHandler := handler.NewRegisterHandler(userService)
	loginHandler := handler.NewLoginHandler(authService)

	http.Handle("/register", registerHandler)
	http.Handle("/login", loginHandler)

	// Защищённые ресурсы
	accountRepo := repository.NewAccountRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	transferRepo := repository.NewTransferRepository(db, txRepo)

	accountService := service.NewAccountService(accountRepo, userRepo, smtpClient, txRepo)
	transferService := service.NewTransferService(transferRepo, accountRepo, userRepo, smtpClient)

	accountHandler := handler.NewAccountHandler(accountService)
	transferHandler := handler.NewTransferHandler(transferService)

	cardRepo := repository.NewCardRepository(db)
	cardService := service.NewCardService(accountRepo, cardRepo, userRepo, txRepo, smtpClient)
	cardHandler := handler.NewCardHandler(cardService)

	creditRepo := repository.NewCreditRepository(db)
	creditService := service.NewCreditService(accountRepo, creditRepo, userRepo, smtpClient)
	creditHandler := handler.NewCreditHandler(creditService)

	http.Handle("/accounts", handler.AuthMiddleware(accountHandler))
	http.Handle("/accounts/", handler.AuthMiddleware(accountHandler))
	http.Handle("/transfers", handler.AuthMiddleware(transferHandler))
	http.Handle("/cards", handler.AuthMiddleware(cardHandler))
	http.Handle("/cards/", handler.AuthMiddleware(cardHandler))
	http.Handle("/credits", handler.AuthMiddleware(creditHandler))
	http.Handle("/credits/", handler.AuthMiddleware(creditHandler))

	// --- Запуск сервера ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Infof("Starting HTTP server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
