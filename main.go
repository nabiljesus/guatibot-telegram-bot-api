package main

import (
    "log"
    "os"
    "fmt"
    "math/rand"
    "time"
    "strings"
    "strconv"
    "sort"
    "net/http"
	"github.com/go-telegram-bot-api/telegram-bot-api"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/sheets/v4"
)

func randomInsult() string {
    rand.Seed(time.Now().Unix())
    reasons := []string{
        "Marico.",
        "Eres gilipollas?",
        "Mamalo.",
        "Uh?",
        "Alto guaraná",
        "Naaa",
        "Marico el que lo lea",
        "Pasa pack.",
    }
    
    return reasons[rand.Intn(len(reasons))]
}

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

func getRangeFromSheet(srv *sheets.Service, spreadsheetId string, cellsRange string) []string {
    resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, cellsRange).Do()
    if err != nil {
            log.Fatalf("Unable to retrieve data from sheet: %v", err)
    }

    var words[]string

    for _, resp := range resp.Values {
        entry := resp[0].(string)
        words = append(words, entry)
    }

    return words
}

func removeDuplicates(words []string) []string {
    keys := make(map[string]bool)

    var filteredWords []string

    for _, word := range words {
        if _, value := keys[word]; !value {
            keys[word] = true
            filteredWords = append(filteredWords, word)
        }
    }

    sort.Strings(filteredWords)

    return filteredWords
}

func addToSheet(words []string) {
	srv := getSheetsService()
    spreadsheetId := os.Getenv("SpreadsheetId")
    cellsRange := "Palabras!A2:A"

	existingWords := getRangeFromSheet(srv, spreadsheetId, cellsRange)
    updatedWords := removeDuplicates(append(existingWords, words...))

    var vr sheets.ValueRange

	for _, word := range updatedWords {
    	var row []interface{}
		row = append(row, word)
        vr.Values = append(vr.Values, row)
	}

    _, err := srv.Spreadsheets.Values.Update(spreadsheetId, cellsRange, &vr).
    								   ValueInputOption("RAW").Do()
    if err != nil {
            log.Fatalf("Unable to write data to sheet: %v", err)
    }

}

func changePercent(newPercentStr string) (string, error) {
    srv := getSheetsService()
    spreadsheetId := os.Getenv("SpreadsheetId")
    cellsRange := "Palabras!D2"

    var vr sheets.ValueRange
    var row []interface{}
    floatPercent, err := strconv.ParseFloat(strings.ReplaceAll(newPercentStr, ",", "."), 64)
    if (err != nil){
        return "", err
    }

    row = append(row, fmt.Sprintf("%.f%%", floatPercent))
    vr.Values = append(vr.Values, row)

    _, err = srv.Spreadsheets.Values.Update(spreadsheetId, cellsRange, &vr).
                                       ValueInputOption("USER_ENTERED").Do()
    return "Todo listo, mano.", err
}

func retrieveWordList() string {
    srv := getSheetsService()
    spreadsheetId := os.Getenv("SpreadsheetId")
    cellsRange := "Palabras!C2"

    wordList := getRangeFromSheet(srv, spreadsheetId, cellsRange)

    if (len(wordList) == 0) {
        return "No hay na', mano."
    } else {
        return wordList[0]
    }
}

func processCommand(message *tgbotapi.Message, response *tgbotapi.MessageConfig){
    var err error

	switch strings.ToLower(message.Command()) {
    	case "addtodibujadera", "add", "añadir", "añade":
			addToSheet(strings.Split(message.CommandArguments(), ","))
    		(*response).Text = "Todo listo, mano."
        case "palabras", "get", "fetch":
            (*response).Text = retrieveWordList()
        case "percent", "porcentaje":
            (*response).Text, err = changePercent(message.CommandArguments())
    	default:
    	   err = fmt.Errorf("Unexpected Command")
	}

    if err != nil {
        log.Fatalf("Command failed with %s", err)
        (*response).Text = "Nolsa, mano."
    }
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BotToken"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
    log.Printf("Authorized on account %s", bot.Self.UserName)

    var updates tgbotapi.UpdatesChannel
    if (strings.Compare(os.Getenv("isLocal"), "true") == 0) {
            bot.RemoveWebhook()
            u := tgbotapi.NewUpdate(0)
            u.Timeout = 60
            updates, err = bot.GetUpdatesChan(u)
    } else {
        _, err = bot.SetWebhook(tgbotapi.NewWebhook(os.Getenv("UrlPath")+bot.Token))
        updates = bot.ListenForWebhook("/update" + bot.Token)
        go http.ListenAndServe(":" + os.Getenv("PORT"), nil)
    }

	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		message := update.Message

		if update.Message == nil { // ignore any non-Message Updates
			message = update.ChannelPost
			if message == nil || !strings.Contains(message.Text, "@guatibot") && !message.IsCommand() {
				continue
			}
		}


		response := tgbotapi.NewMessage(message.Chat.ID, randomInsult())
		response.ReplyToMessageID = message.MessageID

		if message.IsCommand() {
			processCommand(message, &response)
		}

		bot.Send(response)
	}
}