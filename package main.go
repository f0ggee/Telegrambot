package main

import (
	"log"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

// userState хранит "состояние" пользователя (чтобы понимать, что мы у него спрашиваем).
var userState = make(map[int64]string)

// userProfiles хранит профиль для каждого пользователя (по chatID).
var userProfiles = make(map[int64]*user_profile)

type user_profile struct {
	Height int
	Weight int
	Traine int
}

// button – вспомогательный тип для описания кнопок (текст и данные)
type button struct {
	name string
	data string
}

var bot *tgbotapi.BotAPI

func main() {
	// 1. Загружаем токен из .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(".env not loaded (если токен прописан вручную, то всё ок)")
	}

	botToken := os.Getenv("TG_BOT_API")
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Ошибка инициализации Telegram Bot API: %v", err)
	}

	// 2. Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Не удалось получить канал обновлений: %v", err)
	}

	log.Println("Бот запущен...")

	// 3. Главный цикл. Получаем update и обрабатываем
	for update := range updates {
		// 3.1. Если пришёл колбэк (нажатие на инлайн-кнопку)
		if update.CallbackQuery != nil {
			handleCallback(update)
			continue
		}

		// 3.2. Если пришло обычное сообщение
		if update.Message != nil {
			if update.Message.IsCommand() {
				handleCommands(update)
			} else {
				handleMessage(update)
			}
		}
	}
}

// ===========================================
// 1. ФУНКЦИИ ДЛЯ ПОКАЗА МЕНЮ (ИНЛАЙН-КНОПОК)
// ===========================================

// Главное меню
func startMenu() tgbotapi.InlineKeyboardMarkup {
	states := []button{
		{name: "Подсчет калорий", data: "calorie"}, // Пример, пока не реализован
		{name: "Тренировка", data: "traine"},
		{name: "Профиль", data: "profile"},
		{name: "Дневник", data: "dnevnik"},
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, st := range states {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(st.name, st.data),
		)
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// Меню «Тренировка»
func traineMenu() tgbotapi.InlineKeyboardMarkup {
	states := []button{
		{name: "Тренировка: лёгкий уровень", data: "Light"},
		{name: "Тренировка: средний уровень", data: "Midle"},
		{name: "Тренировка: сложный уровень", data: "Hard"},
		{name: "Назад", data: "back"},
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, st := range states {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(st.name, st.data),
		)
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// Пример под-меню для лёгкого уровня (можно упростить или переработать)
func enlightenment() tgbotapi.InlineKeyboardMarkup {
	states := []button{
		{name: "Прокачка бицепса", data: "Bicepslight"},
		{name: "Прокачка передней части руки", data: "handle up"},
		{name: "Прокачка средней части руки", data: "handle middle"},
		{name: "Прокачка задней части руки", data: "handle behind"},
		{name: "Прокачка трицепса", data: "upgrade triceps"},
		{name: "Назад", data: "back2"},
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, st := range states {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(st.name, st.data),
		)
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// Пример под-меню для среднего уровня (можно упростить или переработать)
func enlightenmentMidle() tgbotapi.InlineKeyboardMarkup {
	states := []button{
		{name: "Прокачка бицепса (ср.)", data: "BicepslightM"},
		{name: "Прокачка передней части руки (ср.)", data: "handle upM"},
		{name: "Прокачка средней части руки (ср.)", data: "handle middleM"},
		{name: "Прокачка задней части руки (ср.)", data: "handle behindM"},
		{name: "Прокачка трицепса (ср.)", data: "upgrade tricepsM"},
		{name: "Назад", data: "back3"},
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, st := range states {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(st.name, st.data),
		)
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// Меню «Профиль»
func profileMenu() tgbotapi.InlineKeyboardMarkup {
	states := []button{
		{name: "Введите ваш рост", data: "Ask_height"},
		{name: "Введите ваш вес", data: "Ask_weight"},
		{name: "Ведите сколько вы занимаетесь", data: "Ask_traine"},
		{name: "Показать профиль", data: "Show_profile"},
		{name: "Назад", data: "back"},
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, st := range states {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(st.name, st.data),
		)
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// ===========================================
// 2. ОБРАБОТКА КОЛБЭКОВ (нажатие кнопок)
// ===========================================
func handleCallback(update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID
	messageID := update.CallbackQuery.Message.MessageID

	// Удалим старое сообщение (где были кнопки)
	del := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, _ = bot.Send(del)

	switch data {
	// Главное меню
	case "calorie":
		sendText(chatID, "Здесь будет подсчет калорий (пока не реализовано).")
	case "traine":
		msg := tgbotapi.NewMessage(chatID, "Выберите уровень тренировки:")
		msg.ReplyMarkup = traineMenu()
		sendMessage(msg)
	case "profile":
		msg := tgbotapi.NewMessage(chatID, "Давайте настроим профиль!😃\nЕсли вы не еще делали свой профиль,то пожалуйста,ответьте на следующие вопросы :")
		msg.ReplyMarkup = profileMenu()
		sendMessage(msg)

	case "back":
		// Возвращаемся в главное меню
		msg := tgbotapi.NewMessage(chatID, "Привет! Я ваш помощник-бот. Вот что я умею:\n- 📚 Подсчет калорий.\n- 🏋️‍♂️ Показывать список тренировок.\n- 🧑‍💻 Профиль с вашими данными.\nИспользуйте меню ниже или введите команды вручную. Например:\n/start - Начать работу\n/train - Список тренировок\n/profile - Профиль")
		msg.ReplyMarkup = startMenu()
		sendMessage(msg)

	// Кнопки меню «Тренировка»
	case "Light":
		msg := tgbotapi.NewMessage(chatID, "Вы выбрали лёгкий уровень.")
		msg.ReplyMarkup = enlightenment()
		sendMessage(msg)
	case "Midle":
		msg := tgbotapi.NewMessage(chatID, "Вы выбрали средний уровень.")
		msg.ReplyMarkup = enlightenmentMidle()
		sendMessage(msg)
	case "Hard":
		sendText(chatID, "Вы выбрали сложный уровень.")

	// Кнопки под-меню лёгкого уровня
	case "back2":
		// Вернуться в меню выбора уровня
		msg := tgbotapi.NewMessage(chatID, "Выберите уровень тренировки:")
		msg.ReplyMarkup = traineMenu()
		sendMessage(msg)

	// Можно добавлять обработку "Bicepslight", "handle up", ...
	// пока оставим как пример
	case "Bicepslight":
		sendText(chatID, "Тренировка бицепса (лёгкий уровень).")

	// Кнопки под-меню среднего уровня
	case "back3":
		msg := tgbotapi.NewMessage(chatID, "Выберите уровень тренировки:")
		msg.ReplyMarkup = traineMenu()
		sendMessage(msg)

	// и т.д. для остальных кнопок:
	case "handle upM":
		sendText(chatID, "Ср. уровень, прокачка передней части руки.")
	case "BicepslightM":
		sendText(chatID, "Ср. уровень, прокачка бицепса.")
	// ...

	// Профиль
	case "Ask_height":
		// Ставим состояние, что мы сейчас просим у пользователя Рост
		userState[chatID] = "asking_height"
		sendText(chatID, "Введите ваш рост (например, 170):")

	case "Ask_weight":
		// Ставим состояние, что мы сейчас просим у пользователя Вес
		userState[chatID] = "asking_weight"
		sendText(chatID, "Введите ваш вес (например, 70):")

	case "Ask_traine":
		userState[chatID] = "asking_traine"
		sendText(chatID, "Введите сколько вы занимаетесь в зале ?")

	case "Show_profile":
		// Показываем текущие данные (если есть)
		prof, ok := userProfiles[chatID]
		if !ok {
			sendText(chatID, "Ваш профиль пока пуст.😔 \n Введите рост/вес.")
			return
		}

		message := "Ваш профиль:\n"
		message += "Рост: " + strconv.Itoa(prof.Height) + "\n"
		message += "Вес: " + strconv.Itoa(prof.Weight) + "\n"

		sendText(chatID, message)
	}
}

// ===========================================
// 3. ОБРАБОТКА КОМАНД (например: /start)
// ===========================================
func handleCommands(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	switch update.Message.Command() {
	case "start":
		msg := tgbotapi.NewMessage(chatID, "Привет! Я ваш помощник-бот. Вот что я умею:\n- 📚 Подсчет калорий.\n- 🏋️‍♂️ Показывать список тренировок.\n- 🧑‍💻 Профиль с вашими данными.\n\nИспользуйте меню ниже или введите команды вручную. Например:\n/start - Начать работу\n/train - Список тренировок\n/profile - Профиль")
		msg.ReplyMarkup = startMenu()
		sendMessage(msg)

	case "train":
		msg := tgbotapi.NewMessage(chatID, "Выберите уровень тренировки:")
		msg.ReplyMarkup = traineMenu()
		sendMessage(msg)

	case "profile":
		msg := tgbotapi.NewMessage(chatID, "Давайте настроим профиль!😃\nЕсли вы не еще не делали свой профиль,то пожалуйста,ответьте на следующие вопросы :")
		msg.ReplyMarkup = profileMenu()
		sendMessage(msg)

	default:
		sendText(chatID, "Неизвестная команда: "+update.Message.Command())
	}
}

// ===========================================
// 4. ОБРАБОТКА ОБЫЧНОГО СООБЩЕНИЯ (не команда)
// ===========================================
func handleMessage(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	// Смотрим, что у нас в userState[chatID]
	switch userState[chatID] {

	case "asking_height":
		// Парсим строку в число
		height, err := strconv.Atoi(text)
		if err != nil {
			sendText(chatID, "📝Пожалуйста, введите число (без букв). Попробуйте ещё раз.")
			return
		}

		// Если в карте нет профиля – создадим
		if userProfiles[chatID] == nil {
			userProfiles[chatID] = &user_profile{}
		}
		userProfiles[chatID].Height = height

		stickerMsg := tgbotapi.NewStickerShare(chatID, "CAACAgIAAxkBAAENjCpnjpaR_-huumrAY_ZpSimeXedrOwACgxcAAvr8sUgjW90hqxMV0TYE")
		_, err = bot.Send(stickerMsg)
		if err != nil {
			log.Printf("eror", err)
		}
		time.Sleep(1 * time.Second)

		sendText(chatID, "Ваш рост сохранён!")
		// Сбросим состояние
		userState[chatID] = ""

	case "asking_weight":
		weight, err := strconv.Atoi(text)
		if err != nil {
			sendText(chatID, "📝Пожалуйста, введите число (без букв). Попробуйте ещё раз.")
			return
		}

		if userProfiles[chatID] == nil {
			userProfiles[chatID] = &user_profile{}
		}
		userProfiles[chatID].Weight = weight

		sendText(chatID, "📝Ваш вес сохранён!")
		userState[chatID] = ""

	default:
		// Если мы ни в каком «режиме вопросов» не находимся, можем просто ответить
		sendText(chatID, "Я получил ваше сообщение: "+text)
	}

}

// ===========================================
// 5. ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ОТПРАВКИ
// ===========================================
func sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

func sendMessage(msg tgbotapi.Chattable) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}
