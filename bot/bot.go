package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"tgprogressbot/db"
	"time"
)

var BotToken = os.Getenv("BOT_TOKEN")
var apiURL = "https://api.telegram.org/bot" + BotToken + "/"

var rituals = []string{
	"⚡ Стартуй с ритуала: наполни бутылку, подключи зарядку, надень наушники — мозг поймёт, что пора работать.",
	"🎧 Совмещай приятное с полезным. Учишь Go — пей воду или слушай музыку. Это не отвлекает, а поддерживает.",
	"🧠 Ты не просто решаешь задачи. Ты становишься инженером. Верь в свою новую роль.",
	"🌍 Сиди рядом с теми, кто тоже старается. Библиотека работает как якорь — ты не один.",
	"⏱️ Частота важнее длительности. TikTok по 10 минут 16 раз в день — и день сгорел. Осознавай это.",
	"📍 Cue Ritualization: запускай учёбу с повторяемых действий — и мозг сам войдёт в режим фокуса.",
	"🔄 Identity-Based Habits: говори себе — я не просто читаю, я расту. Это перекалибровывает мотивацию.",
	"💬 Привязывай новую привычку к старой. Например: после душа — 5 минут чтения. Это проще, чем начинать с нуля.",
	"🪩 Соц. среда — мощный рычаг. Учись там, где другие тоже в деле — и ты не расслабишься.",
	"🎯 Главное — не день, а частота. Маленькие шаги каждый день куда важнее, чем редкие подвиги.",
}

var Commands = map[string]func(ChatId int64, args string){
	"/start": handleStart,
	//"/echo":        handleEcho,
	"/help":        handleHelp,
	"/remainder":   handleRemainder,
	"/add":         handleAddWords,
	"/listwords":   handleListWords,
	"/timer":       handleTimer,
	"/about":       handleAbout,
	"/instruction": HandleInstruction,
	//"/test":       // HandleTest,
	"/rituals":   handleRituals,
	"/addgoals":  handleAddGoals,
	"/listgoals": handleListGoals,
}

type Update struct { //main structure of Update
	UpdateID int `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		From *struct {
			Username string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

func GetUpdates(offset int) ([]Update, error) {
	resp, err := http.Get(apiURL + "getUpdates?offset=" + strconv.Itoa(offset))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return result.Result, err
}

func SendMessage(chatID int64, text string) error { //отправляет сообщение
	escaped := url.QueryEscape(text)
	fullURL := apiURL + "sendMessage?chat_id=" + strconv.FormatInt(chatID, 10) + "&text=" + escaped
	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	return nil
}

/*func HandleTest(chatID int64, args string) {
	word, correctAns, count, err := db.GetRandomWord(chatID)
	if err != nil {
		SendMessage(chatID, "You haven't added any word yet. Use /add .")
	}
	SendMessage(chatID, "Write your a translation of "+word)
	if strings.TrimSpace(args) == "" {
		SendMessage(chatID, "What is the meaning of : "+word+"?")
	}
	userans := strings.ToLower(strings.Trimspace(args))
	correctans := strings.Split(strings.ToLower(correctAns), ";")

	Is
	if strings.ToLower(correctAns) == strings.ToLower(ans[1]) {
		SendMessage(chatID, "You've Got it Correct")
		count++
	} else {
		SendMessage(chatID, "You've Got it wrong all of the attemps will be 0 ")
		count = 0
	}
	DB.Exec("UPDATE words SET correct_count = ($1) WHERE user_id = ($2)", count, chatID)
} */

func handleAddGoals(chatId int64, args string) {
	// regex: "goal in quotes" + space + date + space + number
	re := regexp.MustCompile(`"(.+?)"\s+(\d{2}-\d{2}-\d{4})\s+(\d+)`)
	matches := re.FindStringSubmatch(args)

	if len(matches) != 4 {
		SendMessage(chatId, "❗ Формат: /addgoals \"твоя цель\" DD-MM-YYYY N (через сколько дней напоминать(пока что не работает)) обьязательно в ковчках")
		return
	}

	goal := matches[1]
	deadlineStr := matches[2]
	remainderStr := matches[3]

	deadline, err := time.Parse("02-01-2006", deadlineStr)
	if err != nil {
		SendMessage(chatId, "❌ Неверный формат даты. Используй DD-MM-YYYY")
		return
	}

	remainder, err := strconv.Atoi(remainderStr)
	if err != nil {
		SendMessage(chatId, "🔢 Последний параметр должен быть числом — интервал в днях")
		return
	}

	err = db.AddGoal(chatId, goal, deadline, remainder)
	if err != nil {
		SendMessage(chatId, "⚠️ Не удалось сохранить цель: "+err.Error())
		return
	}

	SendMessage(chatId, "🎯 Цель добавлена: \""+goal+"\"\n"+
		"⏳ Дедлайн: "+deadline.Format("02 Jan 2006")+"\n"+
		"🔁 Напоминание каждые "+strconv.Itoa(remainder)+" дней")
}

func handleListGoals(chatId int64, args string) {
	goals, err := db.GetGoals(chatId)
	if err != nil {
		SendMessage(chatId, "faced some problems while getting data")
		fmt.Println(err)
		return
	}

	if len(goals) == 0 {
		SendMessage(chatId, "📭 У тебя пока нет целей. Добавь их с помощью /addgoals")
		return
	}

	message := "Your goals : \n" + strings.Join(goals, "\n")
	SendMessage(chatId, message)

}

func handleAddWords(chatID int64, args string) {
	parts := strings.SplitN(args, " ", 3)

	if len(parts) < 3 {
		SendMessage(chatID, "Not enough arguments please write it in order of  /add {word} {translation} {deadline(1s , 10m , 1h)}")
		return
	}

	word := parts[0]
	translation := parts[1]
	duration, err := time.ParseDuration(parts[2])
	if err != nil {
		SendMessage(chatID, "Wrong format for go u should accomplis like :  10s , 5m , 2h ")
		return
	}
	Deadl := time.Now().Add(duration)
	err = db.AddWords(chatID, word, translation, Deadl)
	if err != nil {
		fmt.Println("faced and error while adding to DB", err)
		SendMessage(chatID, "Have problem while adding your word((")
		return
	}
	SendMessage(chatID, "Word added. I'll remind you after "+duration.String())

	go func() {
		time.Sleep(duration)
		SendMessage(chatID, "It is deadline! you promised yourself to learn about word : "+word)
	}()
}

func handleListWords(chatID int64, args string) {
	words, err := db.GetUserWords(chatID)
	if err != nil {
		SendMessage(chatID, "You haven't saved any words yet . Use /add some.")
		return
	}

	message := "Your vocabluary : \n" + strings.Join(words, "\n")
	SendMessage(chatID, message)
}

func handleStart(chatID int64, args string) {
	SendMessage(chatID, " Hello! I'm your tracker for your progress . I hope I can help to improve you effectivity in any type of spheres :D   use /help to know more about this bot")
}

func handleHelp(chatID int64, args string) {
	message :=
		`There is the list of the commands:
	0. /about (start with it before working with bot)
	0.1 /instruction how to use it in more effective way. 
	1. /start
	2. /add {word} {translation} {time} (Add a new word to learn into vocabulary)
	3. /listwords Shows the every word that you added.
	4. tracker (Shows what you wanted to do) (didn't add yet)
	5. /remainder {time} {thing to remind (write as one string for_example)}
	6. /timer {time} (Tells when the time has gone)
	7./rituals shows some of my rituals to be more focused.
	8./addgoals "{goal}" DD-MM-YYYY random number . You can add your goals here 
	9./listgoals shows your goals and when u wanted to finish them
	DM me if you want to see something new here is my tg @itachi0824
`
	SendMessage(chatID, message)

}

func handleTimer(chatID int64, args string) {
	parts := strings.SplitN(args, " ", 1)
	durationStr := parts[0]

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		SendMessage(chatID, "Wrong format of the time use 1s , 1m , 1h")
	}

	SendMessage(chatID, "Time is setted I'll remind you after "+duration.String())
	go func() {
		time.Sleep(duration)
		SendMessage(chatID, "Time is up!!!!")
	}()
}

func handleRemainder(chatID int64, args string) {
	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		SendMessage(chatID, "SEND 2 message one is  /remainder {time} {message}  so it works properly")
		return
	}

	durationStr := parts[0]
	message := parts[1]

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		SendMessage(chatID, "Wrong format for go u should accomplis like :  10s , 5m , 2h ")
		return
	}

	SendMessage(chatID, "Remainder placed , i'll remind u then!")

	go func() {
		time.Sleep(duration)
		SendMessage(chatID, " Remainding:  "+message)
	}()
}
func handleAbout(chatID int64, args string) {
	message := `*AssistentTracker_bot* — твой цифровой союзник в пути к лучшей версии себя.

	🧠 Функциональность:
	• Добавляй и учи английские слова
	• Ставь цели и дедлайны
	• Получай напоминания
	• Записывай идеи
	• Формируй дисциплину
	
	Главная идея: не просто хранить — а *напоминать*, *возвращать* и *подталкивать* к действиям. 
	Не забывай свои цели — строй себя шаг за шагом.`

	SendMessage(chatID, message)
}

func HandleInstruction(chatID int64, args string) {
	message := `📌 Как использовать AssistentTracker_bot эффективно
	Этот бот — не цель сам по себе, а инструмент, как велосипед для того, кто идёт к вершине.
	Он не сделает работу за тебя, но поможет не сбиться с пути и не забыть, зачем ты начал.

	🔔 Что нужно сделать вначале:
	Включи уведомления в Telegram, чтобы не пропускать напоминания.

	Используй команду /add, чтобы сохранить важные слова или задачи с дедлайнами.

	Ставь реалистичные таймеры — так ты научишься уважать собственные сроки.

	💡 Как использовать бота каждый день:
	Записывай идеи, даже если они кажутся сырыми.

	Напоминай себе о целях с помощью /remainder.

	Используй /tracker и другие команды, чтобы отслеживать прогресс.

	Сформируй привычку возвращаться к своим целям ежедневно — хотя бы на 5 минут.

	🧠 Советы:
	Цели лучше расписывать в Notion или дневнике, а бот использовать как "сторожа", который будет подталкивать.

	Комбинируй инструменты: Notion, ChatGPT, этот бот — каждый покрывает свой кусок. Вместе — это мощь.

	Читайте статьи, изучайте опыт других, формируйте дисциплину — и вы станете на голову выше большинства.`
	SendMessage(chatID, message)
}

func handleRituals(chatId int64, args string) {
	rand.Seed(time.Now().UnixNano())

	index := rand.Intn(len(rituals))
	SendMessage(chatId, rituals[index])

}

//func handleEcho(chatID int64, args string) {
//	SendMessage(chatID, "You said : "+args+"")
//}
