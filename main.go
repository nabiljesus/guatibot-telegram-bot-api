package main

import (
    "log"
    "os"
    "strings"
    "sort"
    "net/http"
	"github.com/go-telegram-bot-api/telegram-bot-api"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/sheets/v4"
)

func getSheetsService() *sheets.Service {
	creds := []byte(os.Getenv("GoogleCreds"))

    // If modifying these scopes, delete your previously saved token.json.
    config, err := google.JWTConfigFromJSON(creds, "https://www.googleapis.com/auth/spreadsheets")
    if err != nil {
            log.Fatalf("Unable to parse client secret file to config: %v", err)
    }
    client := config.Client(oauth2.NoContext)

    srv, err := sheets.New(client)
    if err != nil {
            log.Fatalf("Unable to retrieve Sheets client: %v", err)
    }

    return srv
}

func addToSheet(words []string) {
	srv := getSheetsService()
    //if err != nil {
    //        log.Fatalf("Unable to retrieve Sheets client: %v", err)
    //}

    spreadsheetId := os.Getenv("SpreadsheetId")
    cellsRange := "Palabras!A2:A"

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, cellsRange).Do()
    if err != nil {
            log.Fatalf("Unable to retrieve data from sheet: %v", err)
    }

    keys := make(map[string]bool)

    for _, resp := range resp.Values {
    	entry := resp[0].(string)
        if _, value := keys[entry]; !value {
            keys[entry] = true
            words = append(words, entry)
        }
    }

    if err != nil {
            log.Fatalf("Unable to write data to sheet: %v", err)
    }

    var vr sheets.ValueRange

    var rows [][]interface{}

	for _, word := range words {
        if _, value := keys[word]; !value {
            keys[word] = true
            words = append(words, strings.Title(strings.TrimSpace(word)))
		}
	}

    sort.Strings(words)

	for _, word := range words {
    	var row []interface{}
		row = append(row, word)
    	rows = append(rows, row)
	}

	for _, row := range rows {
		vr.Values = append(vr.Values, row)
	}

    _, err = srv.Spreadsheets.Values.Update(spreadsheetId, cellsRange, &vr).
    								   ValueInputOption("RAW").Do()
    if err != nil {
            log.Fatalf("Unable to write data to sheet: %v", err)
    }

}

func checkCommand(message *tgbotapi.Message, response *tgbotapi.MessageConfig){
	log.Printf("MMGVOOOO\n")
	switch message.Command() {
	case "addToDibujadera":
		if len(message.Text) > 17 {
			addToSheet(strings.Split(message.Text[17:], ","))
		}
		(*response).Text = "Todo listo, mano."
	default:
		log.Printf("COMANDO! %s\n", message.Command())
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BotToken"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(os.Getenv("UrlPath")+bot.Token))
	if err != nil {
		log.Fatal(err)
	}

	updates := bot.ListenForWebhook("/update" + bot.Token)

	go http.ListenAndServe(":" + os.Getenv("PORT"), nil)

	for update := range updates {
		message := update.Message

		if update.Message == nil { // ignore any non-Message Updates
			message = update.ChannelPost
			if !strings.Contains(message.Text, "@guatibot") && !message.IsCommand() {
				continue
			}
		}


		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		response := tgbotapi.NewMessage(message.Chat.ID, "Marico.")
		response.ReplyToMessageID = message.MessageID

		if message.IsCommand() {
			checkCommand(message, &response)
		}

		bot.Send(response)
	}
}