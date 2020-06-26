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

    for _, resp := range resp.Values {
    	words = append(words, resp[0].(string))
    }

    sort.Strings(words)

    if err != nil {
            log.Fatalf("Unable to write data to sheet: %v", err)
    }

    var vr sheets.ValueRange

    var rows [][]interface{}

	for _, word := range words {
    	var row []interface{}
		row = append(row, strings.Title(strings.TrimSpace(word)))
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
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		
		response := "Marico."

		if strings.HasPrefix(update.Message.Text, "/addToDibujadera") {
			addToSheet(strings.Split(update.Message.Text[17:], ","))
			response = "Todo listo, mano."
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}