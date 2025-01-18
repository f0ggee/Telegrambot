package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

// Глобальные переменные
var (
	bot *tgbotapi.BotAPI
)

// Пример структуры для кнопок
type button struct {
	name string
	data string
}

// /Меню профиля
func profile() tgbotapi.InlineKeyboardMarkup {
	states := []button{
		{name: "создать профиль", data: "create"},
		{name: "Посмотреть профиль", data: "check"},
		{name: "Назад", data: "backe"},
	}
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, st := range states {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(st.name, st.data),
		)
		buttons = append(buttons, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// Главное меню (пример)
func startMenu() tgbotapi.InlineKeyboardMarkup {
	states := []button{
		{name: "Подсчет калорий", data: "calorie"},
		{name: "Тренировка", data: "traine"},
		{name: "Профиль", data: "profilee"},
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, st := range states {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(st.name, st.data),
		)
		buttons = append(buttons, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// Меню «Тренировка»
func traineMenu() tgbotapi.InlineKeyboardMarkup {
	states := []button{
		{name: "Тренировка: лёгкий уровень", data: "Light"},
		{name: "Тренировка: средний уровень", data: "Midle"},
		{name: "Тренировка: сложный уровень", data: "Hard"},
		{name: "Назад", data: "back"},
	}

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, st := range states {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(st.name, st.data),
		)
		buttons = append(buttons, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func main() {
	// Загружаем токен из .env (или откуда вам удобно)
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(".env not loaded (it's okay if you have token in another place)")
	}

	botToken := os.Getenv("TG_BOT_API")
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot API: %v", err)
	}

	// Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Failed to get updates channel: %v", err)
	}

	// Главный цикл обработки
	for update := range updates {
		// Обрабатываем колбэки от инлайн-кнопок
		if update.CallbackQuery != nil {
			callbacks(update)
			continue
		}

		// Обрабатываем входящие сообщения

	}
}

// Функция обработки колбэков
func callbacks(update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID
	messageID := update.CallbackQuery.Message.MessageID

	switch data {

	case "create":
		del := tgbotapi.NewDeleteMessage(chatID, messageID)
		bot.Send(del)

		msg := tgbotapi.NewMessage(chatID, "Пока нельзя создать")
		msg.ReplyMarkup = profile()
		bot.Send(msg)

	case "check":
		del := tgbotapi.NewDeleteMessage(chatID, messageID)
		bot.Send(del)

		msg := tgbotapi.NewMessage(chatID, "Профиля неут пока ")
		msg.ReplyMarkup = profile()
		bot.Send(msg)

	case "backe":
		del := tgbotapi.NewDeleteMessage(chatID, messageID)
		bot.Send(del)

		newMenu := startMenu()
		edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, newMenu)
		bot.Send(edit)

	case "traine":
		// Удаляем старое сообщение
		del := tgbotapi.NewDeleteMessage(chatID, messageID)
		bot.Send(del)

		// Выводим меню тренировок
		msg := tgbotapi.NewMessage(chatID, "Это список тренировок по уровням:")
		msg.ReplyMarkup = traineMenu()
		bot.Send(msg)

	case "back":
		// Удаляем старое сообщение (например, меню тренировок)
		del := tgbotapi.NewDeleteMessage(chatID, messageID)
		bot.Send(del)

		// Возвращаемся в главное меню
		msg := tgbotapi.NewMessage(chatID, "Вы в главном меню:")
		msg.ReplyMarkup = startMenu()
		bot.Send(msg)

	case "Light":
		// Логика лёгкого уровня
		sendText(chatID, "Вы выбрали лёгкий уровень")
	case "Midle":
		// Логика среднего уровня
		sendText(chatID, "Вы выбрали средний уровень")
	case "Hard":
		// Логика сложного уровня
		sendText(chatID, "Вы выбрали сложный уровень")
	}
}

// Функция обработки команд /start, /help, /train и т.п.
func commands(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	switch update.Message.Command() {
	case "start":
		msg := tgbotapi.NewMessage(chatID, "Выберите действие")
		msg.ReplyMarkup = startMenu()
		sendMessage(msg)

	case "train":
		msg := tgbotapi.NewMessage(chatID, "Это список тренировок по уровню сложности:")
		msg.ReplyMarkup = traineMenu()
		sendMessage(msg)

	case "profile":
		msg := tgbotapi.NewMessage(chatID, "Выберите действия:")
		msg.ReplyMarkup = profile()
		sendMessage(msg)

	default:
		sendText(chatID, "Неизвестная команда: "+update.Message.Command())
	}
}

// Функция обработки обычного текста (не команды и не колбэка)

// Пример функции суммирования чисел из строки
func sumNumbers(input string) (int, bool) {
	// Заменяем запятые на пробелы, чтобы можно было вводить "1,2,3"
	input = strings.ReplaceAll(input, ",", " ")
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return 0, false
	}
	sum := 0
	for _, p := range parts {
		num, err := strconv.Atoi(p)
		if err != nil {
			return 0, false
		}
		sum += num
	}
	return sum, true
}

// Обёртка для отправки простого текстового сообщения
func sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

// Обёртка для отправки любого Chattable-сообщения
func sendMessage(msg tgbotapi.Chattable) {
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

// Заглушка для сохранения данных (например, в БД)
func SaveNumbers(chatID int64, input string) error {
	// Тут ваша логика сохранения. Пока просто заглушка.
	return nil
}
