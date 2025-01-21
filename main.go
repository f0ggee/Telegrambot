package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func calculateCalories(gender string, weight float64, height float64, age int) float64 {
	if gender == "мужской" {
		return 88.36 + (13.4 * weight) + (4.8 * height) - (5.7 * float64(age))
	}
	return 447.6 + (9.2 * weight) + (3.1 * height) - (4.3 * float64(age))
}

func main() {
	// Создаём бота
	bot, err := tgbotapi.NewBotAPI("7182429562:AAGBcu7cddZF0jAwgVlA4uOzhbRdGt-PO18")
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	log.Printf("Бот авторизован как: %s", bot.Self.UserName)

	// Настройка обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Хранилище данных пользователей
	userData := make(map[int64]map[string]string)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID

		// Проверяем, если пользователь только начинает
		if _, ok := userData[chatID]; !ok {
			userData[chatID] = map[string]string{}
			msg := tgbotapi.NewMessage(chatID, "Привет! Я помогу рассчитать твоё дневное количество калорий. Введи свой вес в кг:")
			bot.Send(msg)
			continue
		}

		// Обработка ввода данных
		userInfo := userData[chatID]

		if _, ok := userInfo["weight"]; !ok {
			weight, err := strconv.ParseFloat(update.Message.Text, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Пожалуйста, введи вес в числовом формате (например: 70.5):")
				bot.Send(msg)
				continue
			}
			userInfo["weight"] = fmt.Sprintf("%.1f", weight)
			msg := tgbotapi.NewMessage(chatID, "Теперь введи свой рост в сантиметрах:")
			bot.Send(msg)
			continue
		}

		if _, ok := userInfo["height"]; !ok {
			height, err := strconv.ParseFloat(update.Message.Text, 64)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Пожалуйста, введи рост в числовом формате (например: 175):")
				bot.Send(msg)
				continue
			}
			userInfo["height"] = fmt.Sprintf("%.1f", height)
			msg := tgbotapi.NewMessage(chatID, "Укажи свой возраст в годах:")
			bot.Send(msg)
			continue
		}

		if _, ok := userInfo["age"]; !ok {
			age, err := strconv.Atoi(update.Message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Пожалуйста, введи возраст в числовом формате (например: 25):")
				bot.Send(msg)
				continue
			}
			userInfo["age"] = strconv.Itoa(age)
			msg := tgbotapi.NewMessage(chatID, "Теперь укажи свой пол (мужской или женский):")
			bot.Send(msg)
			continue
		}

		if _, ok := userInfo["gender"]; !ok {
			gender := strings.ToLower(strings.TrimSpace(update.Message.Text))
			if gender != "мужской" && gender != "женский" {
				msg := tgbotapi.NewMessage(chatID, "Пожалуйста, укажи свой пол: мужской или женский.")
				bot.Send(msg)
				continue
			}
			userInfo["gender"] = gender

			// Все данные получены, расчёт калорий
			weight, _ := strconv.ParseFloat(userInfo["weight"], 64)
			height, _ := strconv.ParseFloat(userInfo["height"], 64)
			age, _ := strconv.Atoi(userInfo["age"])
			gender = userInfo["gender"]

			calories := calculateCalories(gender, weight, height, age)
			result := fmt.Sprintf("Твой базовый обмен веществ (калории в день): %.2f ккал.", calories)

			msg := tgbotapi.NewMessage(chatID, result)
			bot.Send(msg)
			delete(userData, chatID) // Очистка данных пользователя
		}
	}
}
