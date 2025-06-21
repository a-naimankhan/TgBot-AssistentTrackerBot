package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"tgprogressbot/bot"
	"tgprogressbot/db"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ Error loading .env file")
	}
}

func main() {
	if err := db.runMigrations(DB); err != nil {
		log.Fatal("Migration failed:", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Bot is running.")
	})

	log.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))

	err := db.InitDB() // подключение дб
	if err != nil {
		log.Fatalln("found and err while trying to connect with db ", err)
	}

	offset := 0

	//go startHTTPserver()
	for {
		updates, err := bot.GetUpdates(offset)
		time.Sleep(1 * time.Second)
		//проверка того работает ли бот
		if err != nil {
			fmt.Println("Ошибка:", err)
			continue
		}

		for _, update := range updates { //главный луп который помогает с итерацией через апдейты
			//if update.Message == nil || update.Message.Chat.ID == 0 {
			//	continue
			//}
			text := update.Message.Text
			chatID := update.Message.Chat.ID

			username := ""
			if update.Message.From != nil {
				username = update.Message.From.Username
			}
			_ = db.AddUser(chatID, username)

			parts := strings.SplitN(text, " ", 2) //разделение сообщение на /команд и после
			command := parts[0]
			args := ""

			if len(parts) > 1 {
				args = parts[1]
			}

			handler, ok := bot.Commands[command]
			if ok {
				handler(chatID, args)
			} else {
				bot.SendMessage(chatID, "Unknown command. Please try it again or use /help. ")
			}

			fmt.Printf("New message from %d: %s\n", update.Message.Chat.ID, update.Message.Text) //проверка для себя

			offset = update.UpdateID + 1
		}
	}
}
