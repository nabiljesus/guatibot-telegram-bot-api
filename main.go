package main

import (
    "log"
    "os"
    "strings"
    "sort"

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

	//srv, err := sheets.NewService(context.Background(), option.WithAPIKey("AIzaSyCHhmm8OIEOOizb0QvdMZIJleloJ1giRsM"))
    //if err != nil {
    //        log.Fatalf("Unable to retrieve Sheets client: %v", err)
    //}

    // Prints the names and majors of students in a sample spreadsheet:
    // https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
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
    
    sort.Strings(words)

	for _, word := range words {
        if _, value := keys[word]; !value {
            keys[word] = true
	    	var row []interface{}
			row = append(row, strings.Title(strings.TrimSpace(word)))
	    	rows = append(rows, row)
		}
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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		message := update.Message

		if update.Message == nil { // ignore any non-Message Updates
			message = update.ChannelPost
			if !strings.HasPrefix(message.Text, "@guatibot") && !message.IsCommand() {
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