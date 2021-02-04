package main

import (
	"covidgraphs"
	"encoding/json"
	"fmt"
	"github.com/NicoNex/echotron"
	"github.com/robfig/cron"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	nTopRegions      = 10
	botDataDirectory = "/CovidBot"
	imageFolder      = "/plots/"
	logsFolder       = "/logs/"
	botUsername      = "@covidata19bot"
)

var workingDirectory string

// User runtime data struct
type bot struct {
	chatId int64
	echotron.Api
	dailyUpdate             bool     // This is a spoiler for a future feature
	lastButton              string   // Callback of the last pressed button
	lastRegion              string   // Callback of the last pressed button in case it is a region name
	lastProvince            string   // Callback of the last pressed button in case it is a province name
	choicesConfrontoNazione []string // National fields selected for comparison
	choicesConfrontoRegione []string // Regional fields selected for comparison
}

var nationData []covidgraphs.NationData      // National data array
var regionsData []covidgraphs.RegionData     // Regional data array
var provincesData []covidgraphs.ProvinceData // Provincial data array
var datiNote []covidgraphs.NoteData          // Notes array

var natregAttributes = []string{"ricoverati_con_sintomi", "terapia_intensiva", "totale_ospedalizzati",
	"isolamento_domiciliare", "totale_positivi", "nuovi_positivi", "dimessi_guariti", "deceduti",
	"totale_casi", "tamponi"} // National and regional fields names

var reports = []string{"generale"} // Types of reports avvailable

var helpMsg = fmt.Sprintf(`/nazione <code>andamento</code>
per ottenere l'andamento della nazione
/nazione <code>nome_dei_campi</code>
per ottenere un confronto tra campi a tua scelta

/regione <code>nome_regione andamento</code>
per ottenere l'andamento della regione scelta
/regione <code>nome_regione nome_dei_campi</code>
per ottenere un confronto tra campi a tua scelta sulla desiderata

/provincia <code>nome_provincia totale_casi</code>
per ottenere informazioni sul totale dei casi della provincia scelta
/provincia <code>nome_provincia nuovi_positivi</code>
per ottenere informazioni sui nuovi positivi della provincia scelta

/reports <code>[file] nome_report</code>


Dati nazione disponibili:
{<code>%s</code>}

Dati regione disponibili:
{<code>%s</code>}

Report disponibili:{<code>%s</code>}`, strings.Join(natregAttributes, ","),
	strings.Join(natregAttributes, ","), strings.Join(reports, ","))

var mutex = &sync.Mutex{} // Mutex used when updating data from the pcm-dpc repo

// Creates bot data folders if they don't exist
func initFolders() {
	currentPath, _ := os.Getwd()
	workingDirectory = currentPath + botDataDirectory
	os.MkdirAll(workingDirectory, 0755)
	os.MkdirAll(workingDirectory+imageFolder, 0755)
	os.MkdirAll(workingDirectory+logsFolder, 0755)
}

var TOKEN = os.Getenv("CovidBot")

func newBot(chatId int64) echotron.Bot {
	return &bot{
		chatId: chatId,
		Api:    echotron.NewApi(TOKEN),
	}
}

func main() {
	log.SetOutput(os.Stdout)
	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	initFolders()
	updateData(&nationData, &regionsData, &provincesData, &datiNote)()

	// Planning cronjobs to update data from pcm-dpc repo
	var cronjob = cron.New()
	cronjob.AddFunc("TZ=Europe/Rome 15 17 * * *", updateData(&nationData, &regionsData, &provincesData, &datiNote))
	cronjob.AddFunc("TZ=Europe/Rome 40 17 * * *", updateData(&nationData, &regionsData, &provincesData, &datiNote))
	cronjob.AddFunc("TZ=Europe/Rome 00 18 * * *", updateData(&nationData, &regionsData, &provincesData, &datiNote))
	cronjob.AddFunc("TZ=Europe/Rome 05 18 * * *", updateData(&nationData, &regionsData, &provincesData, &datiNote))
	cronjob.Start()

	// Creating bot instance using webhook mode
	dsp := echotron.NewDispatcher(TOKEN, newBot)
	dsp.ListenWebhook("https://hiddenfile.tk:443/bot", 40987)
}

func (b *bot) Update(update *echotron.Update) {
	if update.CallbackQuery == nil {
		keywords := strings.Split(update.Message.Text, " ")
		if keywords[0] == "/start" || keywords[0] == "/start"+botUsername {
			b.sendStart(update)
		} else if keywords[0] == "/help" || keywords[0] == "/help"+botUsername {
			b.sendHelp(update)
		} else if keywords[0] == "/home" || keywords[0] == "/home"+botUsername {
			b.home(update)
		} else if keywords[0] == "/nazione" || keywords[0] == "/nazione"+botUsername {
			b.textNation(update)
		} else if keywords[0] == "/regione" || keywords[0] == "/regione"+botUsername {
			b.textRegion(update)
		} else if keywords[0] == "/provincia" || keywords[0] == "/provincia"+botUsername {
			b.textProvince(update)
		} else if keywords[0] == "/reports" || keywords[0] == "/reports"+botUsername {
			b.textReport(update)
		} else if keywords[0] == "/credits" || keywords[0] == "/credits"+botUsername {
			b.credits(update.Message.Chat.ID)
		}

	} else {
		cq := update.CallbackQuery
		switch strings.ToLower(cq.Data) {
		case "credits":
			b.credits(update.CallbackQuery.Message.Chat.ID)
			b.AnswerCallbackQuery(cq.ID, "Crediti", false)
		case "nuovi casi nazione":
			dirPath := workingDirectory + imageFolder
			title := "Nuovi Positivi"
			var filename string
			var err error

			filename = dirPath + covidgraphs.FilenameCreator(title)
			if !covidgraphs.IsGraphExisting(filename) {
				err, filename = covidgraphs.NuoviPositiviNazione(&nationData, true, title, filename)

				if err != nil {
					log.Println(err)
				}
			}

			buttons, err := b.makeButtons([]string{"Torna alla Home"}, []string{"home"}, 1)
			if err != nil {
				log.Println(err)
				return
			}

			b.SendPhotoWithKeyboard(filename, setCaptionConfrontoNazione(len(nationData)-1, []string{"nuovi_positivi"}), cq.Message.Chat.ID, buttons, echotron.PARSE_HTML)
			b.AnswerCallbackQuery(cq.ID, "Nuovi casi", false)
			b.lastButton = "zonesButtons"
			b.lastRegion = ""
			b.lastProvince = ""
			break
		case "nuovi casi regione":
			regionId, err := covidgraphs.FindFirstOccurrenceRegion(&regionsData, "denominazione_regione", b.lastRegion)
			if err != nil {
				log.Println(err)
				return
			}

			dirPath := workingDirectory + imageFolder
			title := "Nuovi positivi regione " + regionsData[regionId].Denominazione_regione
			var filename string

			filename = dirPath + covidgraphs.FilenameCreator(title)
			if !covidgraphs.IsGraphExisting(filename) {
				err, filename = covidgraphs.VociRegione(&regionsData, []string{"nuovi_positivi"}, 0, regionId, title, filename)

				if err != nil {
					log.Println(err)
				}
			}

			buttons, err := b.makeButtons([]string{"Torna alla Regione", "Torna alla Home"}, []string{b.lastRegion, "home"}, 1)
			if err != nil {
				log.Println(err)
				return
			}
			regionLastId, err := covidgraphs.FindLastOccurrenceRegion(&regionsData, "denominazione_regione", b.lastRegion)
			if err != nil {
				log.Println(err)
				return
			}

			b.SendPhotoWithKeyboard(filename, setCaptionConfrontoRegione(regionLastId, []string{"nuovi_positivi"}), cq.Message.Chat.ID, buttons, echotron.PARSE_HTML)
			b.AnswerCallbackQuery(cq.ID, "Nuovi casi", false)
			b.lastButton = "province"
			b.lastProvince = ""
			break
		case "storico nazione":
			b.lastButton = cq.Data
			b.lastRegion = ""
			b.lastProvince = ""
			break
		case "zonesbuttons":
			buttons, err := b.zonesButtons()
			if err != nil {
				log.Println(err)
			}
			b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
			b.AnswerCallbackQuery(cq.ID, "Zone", false)
			b.lastButton = cq.Data
			b.lastRegion = ""
			b.lastProvince = ""
			break
		case "confronto dati nazione":
			buttonsNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}
			buttonsCallback := make([]string, 0)
			for _, v := range buttonsNames {
				buttonsCallback = append(buttonsCallback, strings.ToLower(v)+" nazione")
			}
			buttonsNames = append(buttonsNames, "Annulla ❌")
			buttonsCallback = append(buttonsCallback, "annulla")
			buttonsNames = append(buttonsNames, "Fatto ✅")
			buttonsCallback = append(buttonsCallback, "fatto")
			buttons, err := b.makeButtons(buttonsNames, buttonsCallback, 2)
			if err != nil {
				log.Println(err)
				return
			}
			b.EditMessageText(cq.Message.Chat.ID, cq.Message.ID, "Seleziona i campi che vuoi mettere a confronto:", buttons, echotron.PARSE_HTML)
			b.AnswerCallbackQuery(cq.ID, "Crea confronto dati nazione", false)
			b.lastButton = "zonesButtons"
			b.lastRegion = ""
			b.lastProvince = ""
			break
		case "confronto dati regione":
			buttonsNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}
			buttonsCallback := make([]string, 0)
			for _, v := range buttonsNames {
				buttonsCallback = append(buttonsCallback, strings.ToLower(v)+" regione")
			}
			buttonsNames = append(buttonsNames, "Annulla ❌")
			buttonsCallback = append(buttonsCallback, "annulla")
			buttonsNames = append(buttonsNames, "Fatto ✅")
			buttonsCallback = append(buttonsCallback, "fatto")
			buttons, err := b.makeButtons(buttonsNames, buttonsCallback, 2)
			if err != nil {
				log.Println(err)
				return
			}
			b.EditMessageText(cq.Message.Chat.ID, cq.Message.ID, "Seleziona i campi che vuoi mettere a confronto:", buttons)
			b.AnswerCallbackQuery(cq.ID, "Crea confronto dati regione", false)
			b.lastButton = "province"
			b.lastProvince = ""
			break
		case "classifica regioni":
			//TODO: grafico a barre classifica
			homeButton, err := b.makeButtons([]string{"Torna alla Home"}, []string{"home"}, 1)
			if err != nil {
				log.Println(err)
				return
			}
			b.SendMessageWithKeyboard(setCaptionTopRegions(), cq.Message.Chat.ID, homeButton, echotron.PARSE_HTML)
			b.AnswerCallbackQuery(cq.ID, "Classifica zonesButtons", false)
			break
		case "classifica province":
			//TODO: grafico a barre classifica
			homeButton, err := b.makeButtons([]string{"Torna alla Home"}, []string{"home"}, 1)
			if err != nil {
				log.Println(err)
				return
			}
			b.SendMessageWithKeyboard(setCaptionTopProvinces(), cq.Message.Chat.ID, homeButton, echotron.PARSE_HTML)
			b.AnswerCallbackQuery(cq.ID, "Classifica province", false)
			break

		case "nord":
			buttons, err := b.nordRegions()
			if err != nil {
				log.Println(err)
			}
			response := b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
			if !response.Ok {
				log.Println(response.Description)
			}
			b.AnswerCallbackQuery(cq.ID, "Nord", false)
			b.lastButton = cq.Data
			b.lastRegion = ""
			b.lastProvince = ""
			break
		case "centro":
			buttons, err := b.centroRegions()
			if err != nil {
				log.Println(err)
			}
			response := b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
			if !response.Ok {
				log.Println(err)
			}
			b.AnswerCallbackQuery(cq.ID, "Centro", false)
			b.lastButton = cq.Data
			b.lastRegion = ""
			b.lastProvince = ""
			break
		case "sud":
			buttons, err := b.sudRegions()
			if err != nil {
				log.Println(err)
			}
			response := b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
			if !response.Ok {
				log.Println(err)
			}
			b.AnswerCallbackQuery(cq.ID, "Sud", false)
			b.lastButton = cq.Data
			b.lastRegion = ""
			b.lastProvince = ""
			break

		case "province":
			provinces := covidgraphs.GetLastProvincesByRegionName(&provincesData, b.lastRegion)
			provincesNames := make([]string, 0)
			for _, v := range *provinces {
				provincesNames = append(provincesNames, v.Denominazione_provincia)
			}
			provincesCallback := make([]string, 0)
			for _, v := range *provinces {
				provincesCallback = append(provincesCallback, strings.ToLower(v.Denominazione_provincia))
			}
			provincesNames = append(provincesNames, "Annulla ❌")
			provincesCallback = append(provincesCallback, "annulla")

			buttons, err := b.makeButtons(provincesNames, provincesCallback, 2)
			if err != nil {
				log.Println(err)
				b.SendMessage("Impossibile reperire il grafico al momento.\nRiprova più tardi.", cq.Message.Chat.ID)
				return
			}

			b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
			b.AnswerCallbackQuery(cq.ID, "Province "+b.lastButton, false)
			b.lastButton = "province"

		case "home":
			b.sendAndamentoNazionale(cq.Message)
			buttons, err := b.mainMenuButtons()
			if err != nil {
				log.Println(err)
			}
			b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
			b.SendMessageWithKeyboard("Scegli un'opzione", cq.Message.Chat.ID, buttons)
			b.AnswerCallbackQuery(cq.ID, "Home", false)
			b.lastButton = cq.Data
			b.lastRegion = ""
			b.lastProvince = ""
			break
		case "annulla":
			b.back(cq)
			break

		case "reports":
			buttonsNames := []string{"Report generale"}
			buttonsCallback := []string{"report generale"}
			buttonsNames = append(buttonsNames, "Annulla ❌")
			buttonsCallback = append(buttonsCallback, "annulla")
			buttons, err := b.makeButtons(buttonsNames, buttonsCallback, 1)
			if err != nil {
				log.Println(err)
				return
			}
			b.SendMessageWithKeyboard("Seleziona un tipo di report:", cq.Message.Chat.ID, buttons)
			b.AnswerCallbackQuery(cq.ID, "Reports", false)
			b.lastButton = "reports"
			break
		case "report generale":
			buttons, err := b.makeButtons([]string{"Genera file", "Torna alla Home"}, []string{"genera_file", "home"}, 1)
			if err != nil {
				log.Println("Errore", err)
				return
			}

			msg := setCaptionAndamentoNazionale() + "\n\n\n" + setCaptionTopRegions() + "\n" + setCaptionTopProvinces()
			b.SendMessageWithKeyboard(msg, cq.Message.Chat.ID, buttons, echotron.PARSE_HTML)
			b.AnswerCallbackQuery(cq.ID, "Report generale", false)
			b.lastButton = "report generale"
		case "genera_file":
			msg := setCaptionAndamentoNazionale() + "\n\n\n" + setCaptionTopRegions() + "\n" + setCaptionTopProvinces()
			switch b.lastButton {
			case "report generale":
				filename := "report generale-" + time.Now().Format("20060102T150405") + ".txt"
				f, err := os.Create(filename)
				if err != nil {
					fmt.Println(err)
					return
				}
				_, err = f.WriteString(msg)
				if err != nil {
					log.Println(err)
					f.Close()
					b.AnswerCallbackQuery(cq.ID, "Si è verificato un errore", false)
					return
				}
				err = f.Close()
				if err != nil {
					log.Println(err)
					b.AnswerCallbackQuery(cq.ID, "Si è verificato un errore", false)
					return
				}
				b.SendDocument(filename, "", cq.Message.Chat.ID)
				b.AnswerCallbackQuery(cq.ID, "Report generato", false)
				err = os.Remove(filename)
				if err != nil {
					log.Println("can't delete file " + filename)
				}
			}
			break

		default:
			if _, err := covidgraphs.FindFirstOccurrenceRegion(&regionsData, "denominazione_regione", cq.Data); err == nil {
				b.caseRegion(cq)
			} else if _, err = covidgraphs.FindFirstOccurrenceProvince(&provincesData, "denominazione_provincia", cq.Data); err == nil {
				b.caseProvince(cq)
			} else if err = b.caseConfrontoRegione(cq); err == nil {
				break
			} else if err = b.caseConfrontoNazione(cq); err == nil {
				break
			} else {
				log.Println("dati callback incorretti")
			}
		}
	}
}

// Handles "start" command
func (b *bot) sendStart(update *echotron.Update) {
	if update.Message.Chat.Type == "private" {
		buttons, err := b.makeButtons([]string{"Credits 🌟", "Vai ai Dati 📊"}, []string{"credits", "home"}, 1)
		if err != nil {
			log.Println(err)
		}

		messageText :=
			`Benvenuto <b>%s</b>!Questo bot mette a disposizione i dati dell'epidemia di Coronavirus in Italia con grafici e numeri.
Puoi seguire i pulsanti per ottenere comodamente le informazioni che desideri

<b><i>oppure</i></b>
Se ti piace digitare puoi usare i seguenti comandi:
%s

Questi comandi possono sempre tornarti utili! Prova ad <b>aggiungere il bot in un gruppo</b> per tenere informate le tue cerchia.

Cominciamo!`

		b.SendMessageWithKeyboard(fmt.Sprintf(messageText, update.Message.User.FirstName, strings.Join(natregAttributes, ","),
			strings.Join(natregAttributes, ","), strings.Join(reports, ","), helpMsg), update.Message.Chat.ID, buttons, echotron.PARSE_HTML)
	} else {
		msg := `
Questo comando non è disponibile nei gruppi.
Digita /help per scoprire i comandi disponibili.
`
		b.SendMessage(msg, update.Message.Chat.ID, echotron.PARSE_HTML)
	}
}

func (b *bot) sendHelp(update *echotron.Update) {
	b.SendMessage(helpMsg, update.Message.Chat.ID, echotron.PARSE_HTML)
}

// Handles "home" command
func (b *bot) home(update *echotron.Update) {
	log.Println("[HOME] Nation: " + fmt.Sprint(nationData) + "\nRegions: " + fmt.Sprint(regionsData) + "\nProvinces: " + fmt.Sprint(provincesData) + "\t" + update.Message.Chat.FirstName + "\n")
	if update.Message.Chat.Type == "private" {
		b.sendAndamentoNazionale(update.Message)
		buttons, err := b.mainMenuButtons()
		if err != nil {
			log.Println(err)
		}
		b.SendMessageWithKeyboard("Scegli un opzione", update.Message.Chat.ID, buttons)
	}
}

// Handles "report" textual command
func (b *bot) textReport(update *echotron.Update) {
	usageMessage := "<b>Uso Corretto del Comando:\n" +
		"/reports <code>[file] nome_report</code>\nDigita /help per visualizzare il manuale."

	message := strings.Replace(update.Message.Text, "/reports ", "", 1)
	var fieldNames []string
	var flagFile bool = false
	tokens := strings.Split(message, " ")
	if len(tokens) < 1 {
		b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
		return
	}
	for i, _ := range tokens {
		tokens[i] = strings.ToLower(tokens[i])
	}
	if tokens[0] == "file" {
		flagFile = true
	}

	i := 0
	if flagFile {
		i = 1
	}
	for ; i < len(tokens); i++ {
		if res := sort.SearchStrings(reports, tokens[i]); res < len(reports) {
			fieldNames = append(fieldNames, tokens[i])
		} else {
			b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
			return
		}
	}
	for _, v := range fieldNames {
		if v == "generale" {
			msg := setCaptionAndamentoNazionale() + "\n\n\n" + setCaptionTopRegions() + "\n" + setCaptionTopProvinces()
			b.SendMessage(msg, update.Message.Chat.ID, echotron.PARSE_HTML)
			if flagFile {
				filename := "report generale-" + time.Now().Format("20060102T150405") + ".txt"
				f, err := os.Create(filename)
				if err != nil {
					log.Println(err)
				}
				_, err = f.WriteString(msg)
				if err != nil {
					log.Println(err)
					f.Close()
				}
				err = f.Close()
				if err != nil {
					log.Println(err)
				}

				b.DeleteMessage(update.Message.Chat.ID, update.Message.ID)
				b.SendDocument(filename, "", update.Message.Chat.ID)
				err = os.Remove(filename)
				if err != nil {
					log.Println("can't delete file " + filename)
				}
			}
		}
	}
}

// Handles "nazione" textual command
func (b *bot) textNation(update *echotron.Update) {
	usageMessage := "<b>Uso Corretto del Comando:\n</b>/nazione <code>andamento</code>\nper ottenere l'andamento della nazione\n" +
		"/nazione <code>nome_dei_campi</code>\nper ottenere un confronto tra campi a tua scelta\nDigita /help per visualizzare il manuale."

	tokens := strings.Fields(update.Message.Text)
	tokens = tokens[1:]

	var fieldNames []string

	if len(tokens) < 1 {
		b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
		return
	}

	sort.Strings(natregAttributes)
	if tokens[0] == "andamento" {
		b.sendAndamentoNazionale(update.Message)
	} else {
		for i := 0; i < len(tokens); i++ {
			if res := sort.SearchStrings(natregAttributes, tokens[i]); res < len(natregAttributes) {
				fieldNames = append(fieldNames, tokens[i])
			} else {
				b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
				return
			}
		}

		titleAttributes := make([]string, 0)
		for i := 0; i < 3; i++ {
			if i < len(fieldNames) {
				titleAttributes = append(titleAttributes, fieldNames[i])
			} else {
				titleAttributes = append(titleAttributes, "")
			}
		}

		dirPath := workingDirectory + imageFolder
		titleForFilename := "Nazione" + fmt.Sprintf("%s_%s_%s", titleAttributes[0], titleAttributes[1], titleAttributes[2])
		title := "Confronto dati nazione"
		var filename string
		var err error

		filename = dirPath + covidgraphs.FilenameCreator(titleForFilename)
		if !covidgraphs.IsGraphExisting(filename) {
			err, filename = covidgraphs.VociNazione(&nationData, fieldNames, 0, title, filename)

			if err != nil {
				log.Println(err)
			}
		}

		if err != nil {
			log.Println(err)
			b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
			return
		}

		b.DeleteMessage(update.Message.Chat.ID, update.Message.ID)
		b.SendPhoto(filename, setCaptionConfrontoNazione(len(nationData)-1, fieldNames), update.Message.Chat.ID, echotron.PARSE_HTML)
	}
}

// Handles "regione" textual command
func (b *bot) textRegion(update *echotron.Update) {
	dirPath := workingDirectory + imageFolder
	usageMessage := "<b>Uso Corretto del Comando:\n</b>/regione <code>nome_regione andamento</code>\nper ottenere l'andamento della regione scelta\n" +
		"/regione <code>nome_regione nome_dei_campi</code>\nper ottenere un confronto tra campi a tua scelta sulla desiderata\nDigita /help per visualizzare il manuale."

	message := strings.Replace(update.Message.Text, "/regione ", "", 1)
	var fieldNames []string
	tokens := strings.Split(message, " ")

	if len(tokens) < 2 {
		b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
		return
	}
	sort.Strings(natregAttributes)
	tokens[0] = strings.Replace(tokens[0], "_", " ", -1)

	regionId, err := covidgraphs.FindLastOccurrenceRegion(&regionsData, "denominazione_regione", tokens[0])
	if err != nil {
		log.Println(err)
		return
	}
	if tokens[1] == "andamento" {
		b.sendAndamentoRegionale(update.Message, regionId)
	} else {
		for i := 1; i < len(tokens); i++ {
			if res := sort.SearchStrings(natregAttributes, tokens[i]); res < len(natregAttributes) {
				fieldNames = append(fieldNames, tokens[i])
			} else {
				return
			}
		}

		regionCode, err := covidgraphs.FindFirstOccurrenceRegion(&regionsData, "denominazione_regione", tokens[0])
		if err != nil {
			log.Println(err)
			b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
			return
		}

		titleAttributes := make([]string, 0)
		for i := 0; i < 3; i++ {
			if i < len(fieldNames) {
				titleAttributes = append(titleAttributes, fieldNames[i])
			} else {
				titleAttributes = append(titleAttributes, "")
			}
		}

		titleForFilename := "Regione" + regionsData[regionCode].Denominazione_regione + fmt.Sprintf("%s_%s_%s", titleAttributes[0], titleAttributes[1], titleAttributes[2])
		title := "Confronto dati regione"
		var filename string

		filename = dirPath + covidgraphs.FilenameCreator(titleForFilename)
		if !covidgraphs.IsGraphExisting(filename) {
			err, filename = covidgraphs.VociRegione(&regionsData, fieldNames, 0, regionCode, title, filename)

			if err != nil {
				log.Println(err)
			}
		}

		if err != nil {
			log.Println(err)
			b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
			return
		}

		b.DeleteMessage(update.Message.Chat.ID, update.Message.ID)
		b.SendPhoto(filename, setCaptionConfrontoRegione(regionId, fieldNames), update.Message.Chat.ID, echotron.PARSE_HTML)
	}
}

// Handles "provincia" textual command
func (b *bot) textProvince(update *echotron.Update) {
	dirPath := workingDirectory + imageFolder
	usageMessage := "<b>Uso Corretto del Comando:\n</b>/provincia <code>nome_provincia totale_casi</code>" +
		"\nper ottenere informazioni sul totale dei casi della provincia scelta\n" +
		"/provincia <code>nome_provincia nuovi_positivi</code>\nper ottenere informazioni sui nuovi positivi della provincia scelta\nDigita /help per visualizzare il manuale."

	message := strings.Replace(update.Message.Text, "/provincia ", "", 1)
	tokens := strings.Split(message, " ")
	if len(tokens) != 2 {
		b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
		return
	}

	tokens[0] = strings.Replace(tokens[0], "_", " ", -1)

	provinceId, err := covidgraphs.FindFirstOccurrenceProvince(&provincesData, "denominazione_provincia", tokens[0])
	if err != nil {
		log.Println(err)
		return
	}
	if tokens[1] == "totale_casi" {
		title := "Totale Contagi " + provincesData[provinceId].Denominazione_provincia
		var filename string

		filename = dirPath + covidgraphs.FilenameCreator(title)
		if !covidgraphs.IsGraphExisting(filename) {
			provinceIndexes := covidgraphs.GetProvinceIndexesByName(&provincesData, tokens[0])
			err, filename = covidgraphs.TotalePositiviProvincia(&provincesData, provinceIndexes, title, filename)

			if err != nil {
				log.Println(err)
			}
		}

		if err != nil {
			log.Println(err)
			b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
			return
		}

		provinceLastId, err := covidgraphs.FindLastOccurrenceProvince(&provincesData, "denominazione_provincia", tokens[0])
		if err != nil {
			log.Println(err)
			b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
			return
		}

		b.SendPhoto(filename, setCaptionConfrontoProvincia(provinceLastId, []string{tokens[1]}), update.Message.Chat.ID, echotron.PARSE_HTML)
	} else if tokens[1] == "nuovi_positivi" {
		title := "Nuovi Positivi " + provincesData[provinceId].Denominazione_provincia
		var filename string

		filename = dirPath + covidgraphs.FilenameCreator(title)
		if !covidgraphs.IsGraphExisting(filename) {
			provinceIndexes := covidgraphs.GetProvinceIndexesByName(&provincesData, tokens[0])
			err, filename = covidgraphs.NuoviPositiviProvincia(&provincesData, provinceIndexes, true, title, filename)

			if err != nil {
				log.Println(err)
			}
		}

		if err != nil {
			log.Println(err)
			b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
			return
		}

		provinceLastId, err := covidgraphs.FindLastOccurrenceProvince(&provincesData, "denominazione_provincia", tokens[0])
		if err != nil {
			log.Println(err)
			b.SendMessage(usageMessage, update.Message.Chat.ID, echotron.PARSE_HTML)
			return
		}

		b.DeleteMessage(update.Message.Chat.ID, update.Message.ID)
		b.SendPhoto(filename, setCaptionConfrontoProvincia(provinceLastId, []string{tokens[1]}), update.Message.Chat.ID, echotron.PARSE_HTML)
	}
}

// Updates data from pcm-dpc repository
func updateData(nazione *[]covidgraphs.NationData, regioni *[]covidgraphs.RegionData, province *[]covidgraphs.ProvinceData, note *[]covidgraphs.NoteData) func() {
	return func() {
		mutex.Lock()

		covidgraphs.DeleteAllPlots(workingDirectory + imageFolder)

		ptrNazione, err := covidgraphs.GetNation()
		if err != nil {
			log.Println("errore nell'aggiornamento dei dati nazione'")
			log.Println(err)
		}
		*nazione = *ptrNazione

		ptrRegioni, err := covidgraphs.GetRegions()
		if err != nil {
			log.Println("errore nell'aggiornamento dei dati zonesButtons'")
		}
		*regioni = *ptrRegioni

		ptrProvince, err := covidgraphs.GetProvinces()
		if err != nil {
			log.Println("errore nell'aggiornamento dei dati province'")
		}
		*province = *ptrProvince

		ptrNote, err := covidgraphs.GetNotes()
		if err != nil {
			log.Println("errore nell'aggiornamento dei dati note'")
		}
		*note = *ptrNote
		mutex.Unlock()
	}
}

// Sends national trend plot and text with related buttons
func (b *bot) sendAndamentoNazionale(message *echotron.Message) {
	dirPath := workingDirectory + imageFolder
	title := "Andamento nazionale"
	var filename string
	var err error

	filename = dirPath + covidgraphs.FilenameCreator(title)
	if !covidgraphs.IsGraphExisting(filename) {
		fields := make([]string, 3)
		fields[0] = "attualmente_positivi"
		fields[1] = "dimessi_guariti"
		fields[2] = "deceduti"
		err, filename = covidgraphs.VociNazione(&nationData, fields, 0, title, filename)

		if err != nil {
			log.Println(err)
			b.SendMessage("Impossibile reperire il grafico al momento.\nRiprova più tardi.", message.Chat.ID)
			return
		}
	}

	b.SendPhoto(filename, setCaptionAndamentoNazionale(), message.Chat.ID, echotron.PARSE_HTML)
}

// Returns the caption for the national trend plot image
func setCaptionAndamentoNazionale() string {
	lastIndex := len(nationData) - 1
	_, nuoviTotale := covidgraphs.CalculateDelta(nationData[lastIndex-1].Totale_positivi, nationData[lastIndex].Totale_positivi)
	_, nuoviGuariti := covidgraphs.CalculateDelta(nationData[lastIndex-1].Dimessi_guariti, nationData[lastIndex].Dimessi_guariti)
	_, nuoviMorti := covidgraphs.CalculateDelta(nationData[lastIndex-1].Deceduti, nationData[lastIndex].Deceduti)
	_, nuoviPositivi := covidgraphs.CalculateDelta(nationData[lastIndex-1].Nuovi_positivi, nationData[lastIndex].Nuovi_positivi)
	data, err := time.Parse("2006-01-02T15:04:05", nationData[lastIndex].Data)
	if err != nil {
		log.Println("error parsing data in setCaptionAndamentoNazionale()")
	}

	msg := "<b>Andamento nazionale " + data.Format("2006-01-02") + "</b>\n\n" +
		"\n<b>Attualmente positivi: </b>" + strconv.Itoa(nationData[lastIndex].Totale_positivi) + " (<i>" + nuoviTotale + "</i>)" +
		"\n<b>Guariti: </b>" + strconv.Itoa(nationData[lastIndex].Dimessi_guariti) + " (<i>" + nuoviGuariti + "</i>)" +
		"\n<b>Morti: </b>" + strconv.Itoa(nationData[lastIndex].Deceduti) + " (<i>" + nuoviMorti + "</i>)" +
		"\n\n<b>Nuovi positivi: </b>" + strconv.Itoa(nationData[lastIndex].Nuovi_positivi) + " (<i>" + nuoviPositivi + "</i>)"

	if nationData[len(nationData)-1].Note_it != "" {
		i, err := covidgraphs.FindFirstOccurrenceNote(&datiNote, "codice", nationData[len(nationData)-1].Note_it)
		if err != nil {
			log.Println("errore nella ricerca della nota col codice indicato")
		} else {
			var campoProvincia string
			if datiNote[i].Provincia != "" {
				campoProvincia = ", " + datiNote[i].Provincia
			}
			var notesField string
			if datiNote[i].Note != "" {
				notesField = ", " + datiNote[i].Note
			}
			msg += "\n\n<b>Note:</b>\n[<i>" + datiNote[i].Tipologia_avviso + "] " + datiNote[i].Regione + campoProvincia + ": " + datiNote[i].Avviso + notesField + "</i>"
		}
	}

	return msg
}

// Creates buttons sets
func (b *bot) makeButtons(buttonsText []string, callbacksData []string, layout int) ([]byte, error) {
	if layout != 1 && layout != 2 {
		return nil, fmt.Errorf("wrong layout")
	}
	if len(buttonsText) != len(callbacksData) {
		return nil, fmt.Errorf("different text and data length")
	}

	buttons := make([]echotron.InlineButton, 0)
	for i, v := range buttonsText {
		buttons = append(buttons, b.InlineKbdBtn(v, "", callbacksData[i]))
	}

	keys := make([]echotron.InlineKbdRow, 0)
	switch layout {
	case 1:
		for i := 0; i < len(buttons); i++ {
			keys = append(keys, echotron.InlineKbdRow{buttons[i]})
		}
		break
	case 2:
		for i := 0; i < len(buttons); i += 2 {
			if i+1 < len(buttons) {
				keys = append(keys, echotron.InlineKbdRow{buttons[i], buttons[i+1]})
			} else {
				keys = append(keys, echotron.InlineKbdRow{buttons[i]})
			}
		}
		break
	}

	inlineKMarkup := b.InlineKbdMarkup(keys...)
	return inlineKMarkup, nil
}

// Returns the caption for the regions top 10
func setCaptionTopRegions() string {
	top := covidgraphs.GetTopTenRegionsTotaleContagi(&regionsData)
	var msg = "<b>Top " + strconv.Itoa(nTopRegions) + " regioni per contagi</b>\n\n"
	for i := 0; i < nTopRegions; i++ {
		msg += "<b>" + strconv.Itoa(i+1) + ". </b>" + (*top)[i].Denominazione_regione + " (<code>" + strconv.Itoa((*top)[i].Totale_casi) + "</code>)\n"
	}

	return msg
}

// Returns the caption for the provinces top 10
func setCaptionTopProvinces() string {
	top := covidgraphs.GetTopTenProvincesTotaleContagi(&provincesData)
	var msg = "<b>Top " + strconv.Itoa(nTopRegions) + " province per contagi</b>\n\n"
	for i := 0; i < nTopRegions; i++ {
		msg += "<b>" + strconv.Itoa(i+1) + ". </b>" + (*top)[i].Denominazione_provincia + " (<code>" + strconv.Itoa((*top)[i].Totale_casi) + "</code>)\n"
	}

	return msg
}

// Recognizes the callback of regions named buttons
func (b *bot) caseRegion(cq *echotron.CallbackQuery) {
	regionIndex, err := covidgraphs.FindLastOccurrenceRegion(&regionsData, "denominazione_regione", cq.Data)
	if err != nil {
		b.AnswerCallbackQuery(cq.ID, "Si è verificato un errore", false)
		return
	}
	b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
	b.sendAndamentoRegionale(cq.Message, regionIndex)
	buttons, err := b.provinceButtons()
	if err != nil {
		log.Println(err)
		return
	}
	b.SendMessageWithKeyboard("Opzioni disponibili:", cq.Message.Chat.ID, buttons)
	b.AnswerCallbackQuery(cq.ID, "Regione "+regionsData[regionIndex].Denominazione_regione, false)
	b.lastButton = cq.Data
	b.lastRegion = cq.Data
	b.lastProvince = ""
}

// Sends a region trend plot and text with related buttons
func (b *bot) sendAndamentoRegionale(message *echotron.Message, regionIndex int) {
	firstRegionIndex, err := covidgraphs.FindFirstOccurrenceRegion(&regionsData, "codice_regione", regionsData[regionIndex].Codice_regione)
	if err != nil {
		b.SendMessage("Impossibile reperire il grafico al momento.\nRiprova più tardi.", message.Chat.ID)
		return
	}

	dirPath := workingDirectory + imageFolder
	title := "Dati regione" + regionsData[firstRegionIndex].Denominazione_regione
	var filename string

	filename = dirPath + covidgraphs.FilenameCreator(title)
	if !covidgraphs.IsGraphExisting(filename) {
		err, filename = covidgraphs.VociRegione(&regionsData, []string{"totale_casi", "dimessi_guariti", "deceduti"}, 0, firstRegionIndex, title, filename)

		if err != nil {
			b.SendMessage("Impossibile reperire il grafico al momento.\nRiprova più tardi.", message.Chat.ID)
			return
		}
	}

	b.SendPhoto(filename, setCaptionRegion(regionIndex), message.Chat.ID, echotron.PARSE_HTML)
}

// Returns the caption for a regional trend plot image
func setCaptionRegion(regionId int) string {
	_, nuoviTotale := covidgraphs.CalculateDelta(regionsData[regionId-21].Totale_casi, regionsData[regionId].Totale_casi)
	_, nuoviGuariti := covidgraphs.CalculateDelta(regionsData[regionId-21].Dimessi_guariti, regionsData[regionId].Dimessi_guariti)
	_, nuoviMorti := covidgraphs.CalculateDelta(regionsData[regionId-21].Deceduti, regionsData[regionId].Deceduti)
	_, nuoviPositivi := covidgraphs.CalculateDelta(regionsData[regionId-21].Nuovi_positivi, regionsData[regionId].Nuovi_positivi)
	_, nuoviRicoveratiConSintomi := covidgraphs.CalculateDelta(regionsData[regionId-21].Ricoverati_con_sintomi, regionsData[regionId].Ricoverati_con_sintomi)
	_, nuoviTerapiaIntensiva := covidgraphs.CalculateDelta(regionsData[regionId-21].Terapia_intensiva, regionsData[regionId].Terapia_intensiva)
	_, nuoviOspedalizzati := covidgraphs.CalculateDelta(regionsData[regionId-21].Totale_ospedalizzati, regionsData[regionId].Totale_ospedalizzati)
	_, nuoviIsolamentoDomiciliare := covidgraphs.CalculateDelta(regionsData[regionId-21].Isolamento_domiciliare, regionsData[regionId].Isolamento_domiciliare)
	_, nuoviTamponi := covidgraphs.CalculateDelta(regionsData[regionId-21].Tamponi, regionsData[regionId].Tamponi)
	data, err := time.Parse("2006-01-02T15:04:05", regionsData[regionId].Data)
	if err != nil {
		log.Println("error parsing data in setCaptionRegion()")
	}

	msg := "<b>Andamento regione " + regionsData[regionId].Denominazione_regione + " " + data.Format("2006-01-02") + "</b>\n\n" +
		"\n<b>Totale positivi: </b>" + strconv.Itoa(regionsData[regionId].Totale_casi) + " (<i>" + nuoviTotale + "</i>)" +
		"\n<b>Guariti: </b>" + strconv.Itoa(regionsData[regionId].Dimessi_guariti) + " (<i>" + nuoviGuariti + "</i>)" +
		"\n<b>Morti: </b>" + strconv.Itoa(regionsData[regionId].Deceduti) + " (<i>" + nuoviMorti + "</i>)" +
		"\n<b>Nuovi positivi: </b>" + strconv.Itoa(regionsData[regionId].Nuovi_positivi) + " (<i>" + nuoviPositivi + "</i>)" +
		"\n\n<b>Ricoverati con sintomi: </b>" + strconv.Itoa(regionsData[regionId].Ricoverati_con_sintomi) + " (<i>" + nuoviRicoveratiConSintomi + "</i>)" +
		"\n<b>Terapia intensiva: </b>" + strconv.Itoa(regionsData[regionId].Terapia_intensiva) + " (<i>" + nuoviTerapiaIntensiva + "</i>)" +
		"\n<b>Totale ospedalizzati: </b>" + strconv.Itoa(regionsData[regionId].Totale_ospedalizzati) + " (<i>" + nuoviOspedalizzati + "</i>)" +
		"\n<b>Isolamento domiciliare: </b>" + strconv.Itoa(regionsData[regionId].Isolamento_domiciliare) + " (<i>" + nuoviIsolamentoDomiciliare + "</i>)" +
		"\n<b>Tamponi effettuati: </b>" + strconv.Itoa(regionsData[regionId].Tamponi) + " (<i>" + nuoviTamponi + "</i>)"

	if regionsData[regionId].Note_it != "" {
		i, err := covidgraphs.FindFirstOccurrenceNote(&datiNote, "codice", regionsData[regionId].Note_it)
		if err != nil {
			log.Println("errore nella ricerca della nota col codice indicato")
		}

		var campoProvincia string
		if datiNote[i].Provincia != "" {
			campoProvincia = ", " + datiNote[i].Provincia
		}
		var notesField string
		if datiNote[i].Note != "" {
			notesField = ", " + datiNote[i].Note
		}
		msg += "\n\n<b>Note:</b>\n[<i>" + datiNote[i].Tipologia_avviso + "] " + datiNote[i].Regione + campoProvincia + ": " + datiNote[i].Avviso + notesField + "</i>"
	}

	return msg
}

// Recognizes the callback of regions named buttons
func (b *bot) caseProvince(cq *echotron.CallbackQuery) {
	provinceIndex, err := covidgraphs.FindLastOccurrenceProvince(&provincesData, "denominazione_provincia", cq.Data)
	if err != nil {
		log.Printf("province not found %v", err)
		b.AnswerCallbackQuery(cq.ID, "Si è verificato un errore", false)
		return
	}
	b.sendAndamentoProvinciale(cq, provinceIndex)
}

// Sends a province trend plot and text with related buttons
func (b *bot) sendAndamentoProvinciale(cq *echotron.CallbackQuery, provinceIndex int) {
	dirPath := workingDirectory + imageFolder
	title := "Totale Contagi " + provincesData[provinceIndex].Denominazione_provincia
	var filename string
	var err error

	filename = dirPath + covidgraphs.FilenameCreator(title)
	if !covidgraphs.IsGraphExisting(filename) {
		provinceIndexes := covidgraphs.GetProvinceIndexesByName(&provincesData, provincesData[provinceIndex].Denominazione_provincia)
		err, filename = covidgraphs.TotalePositiviProvincia(&provincesData, provinceIndexes, title, filename)

		if err != nil {
			b.SendMessage("Impossibile reperire il grafico al momento.\nRiprova più tardi.", cq.Message.Chat.ID)
			return
		}
	}

	buttonsNames := []string{"Torna alla regione", "Torna alla home"}
	callbackNames := []string{b.lastRegion, "home"}
	buttons, err := b.makeButtons(buttonsNames, callbackNames, 1)
	if err != nil {
		log.Println(err)
		return
	}

	b.SendPhoto(filename, setCaptionProvince(provinceIndex), cq.Message.Chat.ID, echotron.PARSE_HTML)
	b.SendMessageWithKeyboard("Opzioni disponibili:", cq.Message.Chat.ID, buttons)
	b.AnswerCallbackQuery(cq.ID, "Regione "+provincesData[provinceIndex].Denominazione_regione, false)
	b.lastButton = cq.Data
	b.lastProvince = cq.Data
}

// Returns the caption for a provincial trend plot image
func setCaptionProvince(provinceId int) string {
	provinceIndexes := covidgraphs.GetProvinceIndexesByName(&provincesData, provincesData[provinceId].Denominazione_provincia)
	todayIndex := (*provinceIndexes)[len(*provinceIndexes)-1]
	yesterdayIndex := (*provinceIndexes)[len(*provinceIndexes)-2]
	_, nuoviTotale := covidgraphs.CalculateDelta(provincesData[yesterdayIndex].Totale_casi, provincesData[todayIndex].Totale_casi)
	_, nuoviPositivi := covidgraphs.CalculateDelta(provincesData[yesterdayIndex].NuoviCasi, provincesData[todayIndex].NuoviCasi)
	data, err := time.Parse("2006-01-02T15:04:05", provincesData[provinceId].Data)
	if err != nil {
		log.Println("error parsing data in setCaptionAndamentoNazionale()")
	}

	msg := "<b>Andamento provincia di " + provincesData[provinceId].Denominazione_provincia + " " + data.Format("2006-01-02") + "</b>\n\n" +
		"\n<b>Totale positivi: </b>" + strconv.Itoa(provincesData[provinceId].Totale_casi) + " (<i>" + nuoviTotale + "</i>)" +
		"\n\n<b>Nuovi positivi: </b>" + strconv.Itoa(provincesData[provinceId].NuoviCasi) + " (<i>" + nuoviPositivi + "</i>)"

	if provincesData[provinceId].Note_it != "" {
		i, err := covidgraphs.FindFirstOccurrenceNote(&datiNote, "codice", provincesData[provinceId].Note_it)
		if err != nil {
			log.Println("errore nella ricerca della nota col codice indicato")
		}

		var campoProvincia string
		if datiNote[i].Provincia != "" {
			campoProvincia = ", " + datiNote[i].Provincia
		}
		var notesField string
		if datiNote[i].Note != "" {
			notesField = ", " + datiNote[i].Note
		}
		msg += "\n\n<b>Note:</b>\n[<i>" + datiNote[i].Tipologia_avviso + "] " + datiNote[i].Regione + campoProvincia + ": " + datiNote[i].Avviso + notesField
	}

	return msg
}

// Handles "Annulla" button callback to go back according to the current context
func (b *bot) back(cq *echotron.CallbackQuery) {
	writeOperation(cq)
	switch b.lastButton {
	case "zonesButtons":
		buttons, err := b.mainMenuButtons()
		if err != nil {
			log.Println(err)
			return
		}
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Annulla", false)
		b.choicesConfrontoRegione = make([]string, 0)
		b.choicesConfrontoNazione = make([]string, 0)
		break
	case "province":
		buttons, err := b.provinceButtons()
		if err != nil {
			log.Println(err)
			return
		}
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Annulla", false)
		b.lastButton = "province"
		b.choicesConfrontoRegione = make([]string, 0)
		b.choicesConfrontoNazione = make([]string, 0)
		break
	case "nord":
		buttons, err := b.zonesButtons()
		if err != nil {
			log.Println(err)
			return
		}
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Annulla", false)
		b.lastButton = "zonesButtons"
		break
	case "centro":
		buttons, err := b.zonesButtons()
		if err != nil {
			log.Println(err)
			return
		}
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Annulla", false)
		b.lastButton = "zonesButtons"
		break
	case "sud":
		buttons, err := b.zonesButtons()
		if err != nil {
			log.Println(err)
			return
		}
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Annulla", false)
		b.lastButton = "zonesButtons"
		break
	case "reports":
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		b.lastButton = ""
		b.lastRegion = ""
		b.lastProvince = ""
	}
}

// Creates the main menu buttons set
func (b *bot) mainMenuButtons() ([]byte, error) {
	//buttonsNames := []string{"Storico 🕑", "Regioni", "Vai a regione ➡️", "Vai a provincia ➡️", "Crea confronto su dati nazione 📈", "Classifica regioni 🏅", "Classifica province 🏅"}
	buttonsNames := []string{"Nuovi casi 🆕", "Regioni", "Crea confronto su dati nazione 📈", "Classifica regioni 🏅", "Classifica province 🏅", "Reports 📃"}
	//callbackData := []string{"storico nazione", "zonesButtons", "vai a regione", "vai a provincia", "crea confronto su dati nazione", "classifica regioni", "classifica province"}
	callbackData := []string{"nuovi casi nazione", "zonesButtons", "confronto dati nazione", "classifica regioni", "classifica province", "reports"}
	buttons, err := b.makeButtons(buttonsNames, callbackData, 1)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	b.choicesConfrontoNazione = make([]string, 0)
	b.choicesConfrontoRegione = make([]string, 0)
	return buttons, nil
}

// Creates zones buttons set
func (b *bot) zonesButtons() ([]byte, error) {
	zones := []string{"Nord", "Centro", "Sud"}
	zonesCallback := make([]string, 0)
	for _, v := range zones {
		zonesCallback = append(zonesCallback, strings.ToLower(v))
	}
	zones = append(zones, "Annulla ❌")
	zonesCallback = append(zonesCallback, "annulla")
	buttons, err := b.makeButtons(zones, zonesCallback, 1)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return buttons, nil
}

// Creates provinces buttons set
func (b *bot) provinceButtons() ([]byte, error) {
	buttonsNames := []string{"Nuovi casi 🆕", "Province della regione", "Confronto dati regione 📈", "Torna alla home"}
	callbackNames := []string{"nuovi casi regione", "province", "confronto dati regione", "home"}
	buttons, err := b.makeButtons(buttonsNames, callbackNames, 1)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	b.choicesConfrontoRegione = make([]string, 0)
	return buttons, nil
}

// Creates northern regions buttons set
func (b *bot) nordRegions() ([]byte, error) {
	regions := covidgraphs.GetNordRegionsNamesList()
	regionsCallback := make([]string, 0)
	for _, v := range regions {
		regionsCallback = append(regionsCallback, strings.ToLower(v))
	}
	regions = append(regions, "Annulla ❌")
	regionsCallback = append(regionsCallback, "annulla")
	buttons, err := b.makeButtons(regions, regionsCallback, 2)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return buttons, nil
}

// Creates central regions buttons set
func (b *bot) centroRegions() ([]byte, error) {
	regions := covidgraphs.GetCentroRegionsNamesList()
	regionsCallback := make([]string, 0)
	for _, v := range regions {
		regionsCallback = append(regionsCallback, strings.ToLower(v))
	}
	regions = append(regions, "Annulla ❌")
	regionsCallback = append(regionsCallback, "annulla")
	buttons, err := b.makeButtons(regions, regionsCallback, 2)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return buttons, nil
}

// Creates southern regions buttons set
func (b *bot) sudRegions() ([]byte, error) {
	regions := covidgraphs.GetSudRegionsNamesList()
	regionsCallback := make([]string, 0)
	for _, v := range regions {
		regionsCallback = append(regionsCallback, strings.ToLower(v))
	}
	regions = append(regions, "Annulla ❌")
	regionsCallback = append(regionsCallback, "annulla")
	buttons, err := b.makeButtons(regions, regionsCallback, 2)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return buttons, nil
}

// Handles "Confronto dati regione" selected fields
func (b *bot) caseConfrontoRegione(cq *echotron.CallbackQuery) error {
	buttonsNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}
	buttonsCallback := make([]string, 0)
	for _, v := range buttonsNames {
		buttonsCallback = append(buttonsCallback, strings.ToLower(v)+" regione")
	}

	switch cq.Data {
	case "ricoverati con sintomi regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "ricoverati con sintomi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !b.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "terapia intensiva regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "terapia intensiva")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !b.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale ospedalizzati regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "totale ospedalizzati")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !b.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "isolamento domiciliare regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "isolamento domiciliare")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !b.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "attualmente positivi regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "attualmente positivi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !b.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "nuovi positivi regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "nuovi positivi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !b.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "dimessi guariti regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "dimessi guariti")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !b.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "deceduti regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "deceduti")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !b.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale casi regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "totale casi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !b.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "tamponi regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "tamponi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !b.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "fatto regione":
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		b.sendConfrontoDatiRegione(cq)
		b.choicesConfrontoRegione = make([]string, 0)
		break
	default:
		return fmt.Errorf("not a confronto regioni case")
	}

	return nil
}

// Sends a plot with a caption containing a comparison with the selected regional fields
func (b *bot) sendConfrontoDatiRegione(cq *echotron.CallbackQuery) {
	snakeCaseChoices := make([]string, 0)
	for _, v := range b.choicesConfrontoRegione {
		snakeCaseChoices = append(snakeCaseChoices, strings.Replace(v, " ", "_", -1))
	}
	b.choicesConfrontoRegione = snakeCaseChoices

	regionId, err := covidgraphs.FindFirstOccurrenceRegion(&regionsData, "denominazione_regione", b.lastRegion)
	if err != nil {
		log.Println(err)
		return
	}

	sortedChoices := b.getSortedChoicesConfrontoRegione()

	titleAttributes := make([]string, 0)
	for i := 0; i < 3; i++ {
		if i < len(sortedChoices) {
			titleAttributes = append(titleAttributes, sortedChoices[i])
		} else {
			titleAttributes = append(titleAttributes, "")
		}
	}

	dirPath := workingDirectory + imageFolder
	titleForFilename := "Regione" + regionsData[regionId].Denominazione_regione + fmt.Sprintf("%s_%s_%s", titleAttributes[0], titleAttributes[1], titleAttributes[2])
	var filename string
	title := "Confronto dati regione " + regionsData[regionId].Denominazione_regione

	filename = dirPath + covidgraphs.FilenameCreator(titleForFilename)
	if !covidgraphs.IsGraphExisting(filename) {
		err, filename = covidgraphs.VociRegione(&regionsData, b.choicesConfrontoRegione, 0, regionId, title, filename)

		if err != nil {
			log.Println(err)
		}
	}

	buttonsNames := []string{"Torna alla regione", "Torna alla home"}
	callbackNames := []string{b.lastRegion, "home"}
	buttons, err := b.makeButtons(buttonsNames, callbackNames, 1)
	if err != nil {
		log.Println(err)
		return
	}
	regionLastId, err := covidgraphs.FindLastOccurrenceRegion(&regionsData, "denominazione_regione", regionsData[regionId].Denominazione_regione)
	if err != nil {
		log.Println(err)
		return
	}

	b.SendPhoto(filename, setCaptionConfrontoRegione(regionLastId, b.choicesConfrontoRegione), cq.Message.Chat.ID, echotron.PARSE_HTML)
	b.SendMessageWithKeyboard("Opzioni disponibili:", cq.Message.Chat.ID, buttons)
	b.AnswerCallbackQuery(cq.ID, "Confronto effettuato", false)
}

// Returns the caption for the requested regional fields comparison plot
func setCaptionConfrontoRegione(regionId int, fieldsNames []string) string {
	_, nuoviTotale := covidgraphs.CalculateDelta(regionsData[regionId-21].Totale_casi, regionsData[regionId].Totale_casi)
	_, nuoviGuariti := covidgraphs.CalculateDelta(regionsData[regionId-21].Dimessi_guariti, regionsData[regionId].Dimessi_guariti)
	_, nuoviMorti := covidgraphs.CalculateDelta(regionsData[regionId-21].Deceduti, regionsData[regionId].Deceduti)
	_, nuoviTotalePositivi := covidgraphs.CalculateDelta(regionsData[regionId-21].Totale_positivi, regionsData[regionId].Totale_positivi)
	_, nuoviPositivi := covidgraphs.CalculateDelta(regionsData[regionId-21].Nuovi_positivi, regionsData[regionId].Nuovi_positivi)
	_, nuoviRicoveratiConSintomi := covidgraphs.CalculateDelta(regionsData[regionId-21].Ricoverati_con_sintomi, regionsData[regionId].Ricoverati_con_sintomi)
	_, nuoviTerapiaIntensiva := covidgraphs.CalculateDelta(regionsData[regionId-21].Terapia_intensiva, regionsData[regionId].Terapia_intensiva)
	_, nuoviOspedalizzati := covidgraphs.CalculateDelta(regionsData[regionId-21].Totale_ospedalizzati, regionsData[regionId].Totale_ospedalizzati)
	_, nuoviIsolamentoDomiciliare := covidgraphs.CalculateDelta(regionsData[regionId-21].Isolamento_domiciliare, regionsData[regionId].Isolamento_domiciliare)
	_, nuoviTamponi := covidgraphs.CalculateDelta(regionsData[regionId-21].Tamponi, regionsData[regionId].Tamponi)
	data, err := time.Parse("2006-01-02T15:04:05", regionsData[regionId].Data)
	if err != nil {
		log.Println("error parsing data in region caption")
	}

	msg := "<b>Andamento regione " + regionsData[regionId].Denominazione_regione + " " + data.Format("2006-01-02") + "</b>\n"
	for _, v := range fieldsNames {
		if v == "totale_casi" {
			msg += "\n<b>Totale positivi: </b>" + strconv.Itoa(regionsData[regionId].Totale_casi) + " (<i>" + nuoviTotale + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "dimessi_guariti" {
			msg += "\n<b>Guariti: </b>" + strconv.Itoa(regionsData[regionId].Dimessi_guariti) + " (<i>" + nuoviGuariti + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "deceduti" {
			msg += "\n<b>Morti: </b>" + strconv.Itoa(regionsData[regionId].Deceduti) + " (<i>" + nuoviMorti + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "attualmente_positivi" {
			msg += "\n<b>Attualmente positivi: </b>" + strconv.Itoa(regionsData[regionId].Totale_positivi) + " (<i>" + nuoviTotalePositivi + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "nuovi_positivi" {
			msg += "\n<b>Nuovi positivi: </b>" + strconv.Itoa(regionsData[regionId].Nuovi_positivi) + " (<i>" + nuoviPositivi + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "ricoverati_con_sintomi" {
			msg += "\n<b>Ricoverati con sintomi: </b>" + strconv.Itoa(regionsData[regionId].Ricoverati_con_sintomi) + " (<i>" + nuoviRicoveratiConSintomi + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "terapia_intensiva" {
			msg += "\n<b>Terapia intensiva: </b>" + strconv.Itoa(regionsData[regionId].Terapia_intensiva) + " (<i>" + nuoviTerapiaIntensiva + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "totale_ospedalizzati" {
			msg += "\n<b>Totale ospedalizzati: </b>" + strconv.Itoa(regionsData[regionId].Totale_ospedalizzati) + " (<i>" + nuoviOspedalizzati + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "isolamento_domiciliare" {
			msg += "\n<b>Isolamento domiciliare: </b>" + strconv.Itoa(regionsData[regionId].Isolamento_domiciliare) + " (<i>" + nuoviIsolamentoDomiciliare + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "tamponi" {
			msg += "\n<b>Tamponi effettuati: </b>" + strconv.Itoa(regionsData[regionId].Tamponi) + " (<i>" + nuoviTamponi + "</i>)"
		}
	}

	return msg
}

// Checks if a string is found in national fields selected for comparison
func (b *bot) isStringFoundInRegionChoices(str string) bool {
	for _, j := range b.choicesConfrontoRegione {
		if j == strings.ToLower(str) {
			return true
		}
	}
	return false
}

// Handles "Confronto dati nazione" selected fields
func (b *bot) caseConfrontoNazione(cq *echotron.CallbackQuery) error {
	buttonsNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}
	buttonsCallback := make([]string, 0)
	for _, v := range buttonsNames {
		buttonsCallback = append(buttonsCallback, strings.ToLower(v)+" nazione")
	}

	switch cq.Data {
	case "ricoverati con sintomi nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "ricoverati con sintomi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !b.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "terapia intensiva nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "terapia intensiva")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !b.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale ospedalizzati nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "totale ospedalizzati")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !b.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "isolamento domiciliare nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "isolamento domiciliare")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !b.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "attualmente positivi nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "attualmente positivi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !b.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "nuovi positivi nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "nuovi positivi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !b.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "dimessi guariti nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "dimessi guariti")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !b.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "deceduti nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "deceduti")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !b.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale casi nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "totale casi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !b.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "tamponi nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "tamponi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !b.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !b.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := b.makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "fatto nazione":
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		b.sendConfrontoDatiNazione(cq)
		b.choicesConfrontoNazione = make([]string, 0)
		break
	default:
		return fmt.Errorf("not a confronto nazione case")
	}

	return nil
}

// Sends a plot with a caption containing a comparison with the selected national fields
func (b *bot) sendConfrontoDatiNazione(cq *echotron.CallbackQuery) {
	snakeCaseChoices := make([]string, 0)
	for _, v := range b.choicesConfrontoNazione {
		snakeCaseChoices = append(snakeCaseChoices, strings.Replace(v, " ", "_", -1))
	}
	b.choicesConfrontoNazione = snakeCaseChoices

	sortedChoices := b.getSortedChoicesConfrontoNazione()

	titleAttributes := make([]string, 0)
	for i := 0; i < 3; i++ {
		if i < len(sortedChoices) {
			titleAttributes = append(titleAttributes, sortedChoices[i])
		} else {
			titleAttributes = append(titleAttributes, "")
		}
	}

	dirPath := workingDirectory + imageFolder
	titleForFilename := "Nazione" + fmt.Sprintf("%s_%s_%s", titleAttributes[0], titleAttributes[1], titleAttributes[2])
	var filename string
	var err error
	title := "Confronto dati nazione"

	filename = dirPath + covidgraphs.FilenameCreator(titleForFilename)
	if !covidgraphs.IsGraphExisting(filename) {
		err, filename = covidgraphs.VociNazione(&nationData, b.choicesConfrontoNazione, 0, title, filename)

		if err != nil {
			log.Println(err)
		}
	}

	buttonsNames := []string{"Torna alla home"}
	callbackNames := []string{"home"}
	buttons, err := b.makeButtons(buttonsNames, callbackNames, 1)
	if err != nil {
		log.Println(err)
		return
	}

	b.SendPhoto(filename, setCaptionConfrontoNazione(len(nationData)-1, b.choicesConfrontoNazione), cq.Message.Chat.ID, echotron.PARSE_HTML)
	b.SendMessageWithKeyboard("Opzioni disponibili:", cq.Message.Chat.ID, buttons)
	b.AnswerCallbackQuery(cq.ID, "Confronto effettuato", false)
}

// Returns the caption for the requested national fields comparison plot
func setCaptionConfrontoNazione(nationId int, fieldsNames []string) string {
	_, nuoviTotale := covidgraphs.CalculateDelta(nationData[nationId-1].Totale_casi, nationData[nationId].Totale_casi)
	_, nuoviGuariti := covidgraphs.CalculateDelta(nationData[nationId-1].Dimessi_guariti, nationData[nationId].Dimessi_guariti)
	_, nuoviMorti := covidgraphs.CalculateDelta(nationData[nationId-1].Deceduti, nationData[nationId].Deceduti)
	_, nuoviTotalePositivi := covidgraphs.CalculateDelta(nationData[nationId-1].Totale_positivi, nationData[nationId].Totale_positivi)
	_, nuoviPositivi := covidgraphs.CalculateDelta(nationData[nationId-1].Nuovi_positivi, nationData[nationId].Nuovi_positivi)
	_, nuoviRicoveratiConSintomi := covidgraphs.CalculateDelta(nationData[nationId-1].Ricoverati_con_sintomi, nationData[nationId].Ricoverati_con_sintomi)
	_, nuoviTerapiaIntensiva := covidgraphs.CalculateDelta(nationData[nationId-1].Terapia_intensiva, nationData[nationId].Terapia_intensiva)
	_, nuoviOspedalizzati := covidgraphs.CalculateDelta(nationData[nationId-1].Totale_ospedalizzati, nationData[nationId].Totale_ospedalizzati)
	_, nuoviIsolamentoDomiciliare := covidgraphs.CalculateDelta(nationData[nationId-1].Isolamento_domiciliare, nationData[nationId].Isolamento_domiciliare)
	_, nuoviTamponi := covidgraphs.CalculateDelta(nationData[nationId-1].Tamponi, nationData[nationId].Tamponi)
	data, err := time.Parse("2006-01-02T15:04:05", nationData[nationId].Data)
	if err != nil {
		log.Println("error parsing data in nation caption")
	}

	msg := "<b>Andamento nazione " + data.Format("2006-01-02") + "</b>\n"
	for _, v := range fieldsNames {
		if v == "totale_casi" {
			msg += "\n<b>Totale positivi: </b>" + strconv.Itoa(nationData[nationId].Totale_casi) + " (<i>" + nuoviTotale + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "dimessi_guariti" {
			msg += "\n<b>Guariti: </b>" + strconv.Itoa(nationData[nationId].Dimessi_guariti) + " (<i>" + nuoviGuariti + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "deceduti" {
			msg += "\n<b>Morti: </b>" + strconv.Itoa(nationData[nationId].Deceduti) + " (<i>" + nuoviMorti + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "attualmente_positivi" {
			msg += "\n<b>Attualmente positivi: </b>" + strconv.Itoa(nationData[nationId].Totale_positivi) + " (<i>" + nuoviTotalePositivi + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "nuovi_positivi" {
			msg += "\n<b>Nuovi positivi: </b>" + strconv.Itoa(nationData[nationId].Nuovi_positivi) + " (<i>" + nuoviPositivi + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "ricoverati_con_sintomi" {
			msg += "\n<b>Ricoverati con sintomi: </b>" + strconv.Itoa(nationData[nationId].Ricoverati_con_sintomi) + " (<i>" + nuoviRicoveratiConSintomi + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "terapia_intensiva" {
			msg += "\n<b>Terapia intensiva: </b>" + strconv.Itoa(nationData[nationId].Terapia_intensiva) + " (<i>" + nuoviTerapiaIntensiva + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "totale_ospedalizzati" {
			msg += "\n<b>Totale ospedalizzati: </b>" + strconv.Itoa(nationData[nationId].Totale_ospedalizzati) + " (<i>" + nuoviOspedalizzati + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "isolamento_domiciliare" {
			msg += "\n<b>Isolamento domiciliare: </b>" + strconv.Itoa(nationData[nationId].Isolamento_domiciliare) + " (<i>" + nuoviIsolamentoDomiciliare + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "tamponi" {
			msg += "\n<b>Tamponi effettuati: </b>" + strconv.Itoa(nationData[nationId].Tamponi) + " (<i>" + nuoviTamponi + "</i>)"
		}
	}

	return msg
}

// Checks if a string is found in national fields selected for comparison
func (b *bot) isStringFoundInNationChoices(str string) bool {
	for _, j := range b.choicesConfrontoNazione {
		if j == strings.ToLower(str) {
			return true
		}
	}
	return false
}

// Returns a caption with the selected province fields data
func setCaptionConfrontoProvincia(provinceId int, fieldsNames []string) string {
	provinceIndexes := covidgraphs.GetProvinceIndexesByName(&provincesData, provincesData[provinceId].Denominazione_provincia)
	todayIndex := (*provinceIndexes)[len(*provinceIndexes)-1]
	yesterdayIndex := (*provinceIndexes)[len(*provinceIndexes)-2]
	_, nuoviTotale := covidgraphs.CalculateDelta(provincesData[yesterdayIndex].Totale_casi, provincesData[todayIndex].Totale_casi)
	_, nuoviPositivi := covidgraphs.CalculateDelta(provincesData[yesterdayIndex].NuoviCasi, provincesData[todayIndex].NuoviCasi)
	data, err := time.Parse("2006-01-02T15:04:05", provincesData[provinceId].Data)
	if err != nil {
		log.Println("error parsing data for province caption")
	}

	msg := "<b>Andamento provincia di " + provincesData[provinceId].Denominazione_provincia + " " + data.Format("2006-01-02") + "</b>\n"
	for _, v := range fieldsNames {
		if v == "totale_casi" {
			msg += "\n<b>Totale positivi: </b>" + strconv.Itoa(provincesData[provinceId].Totale_casi) + " (<i>" + nuoviTotale + "</i>)"
		}
	}
	for _, v := range fieldsNames {
		if v == "nuovi_positivi" {
			msg += "\n<b>Nuovi positivi: </b>" + strconv.Itoa(provincesData[provinceId].NuoviCasi) + " (<i>" + nuoviPositivi + "</i>)"
		}
	}

	return msg
}

// Logs operations to a file named with telegram username/name_surname
func writeOperation(cq *echotron.CallbackQuery) {
	data, err := json.Marshal(cq)
	if err != nil {
		log.Println(err)
		return
	}

	dirPath := workingDirectory + logsFolder
	filename := dirPath + cq.Message.Chat.Username + ".txt"
	if cq.Message.Chat.Username == "" {
		filename = dirPath + cq.Message.Chat.FirstName + "_" + cq.Message.Chat.LastName + ".txt"
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		return
	}

	dataString := time.Now().Format("2006-01-02T15:04:05") + string(data[:])
	_, err = f.WriteString(dataString)
	if err != nil {
		log.Println(err)
		return
	}
	err = f.Close()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(dataString)
}

// Sends credits message
func (b *bot) credits(chatId int64) {
	buttons, err := b.makeButtons([]string{"Torna alla Home"}, []string{"home"}, 1)
	if err != nil {
		log.Println(err)
	}
	b.SendMessageWithKeyboard("🤖 Bot creato da @GiovanniRanaTortello\n😺 GitHub: https://github.com/DarkFighterLuke\n"+
		"\n🌐 Proudly hosted on Raspberry Pi 3", chatId, buttons, echotron.PARSE_HTML)
}

// Sorts regional fields selected for comparison
func (b *bot) getSortedChoicesConfrontoRegione() []string {
	tempChoices := make([]string, len(b.choicesConfrontoRegione))
	copy(tempChoices, b.choicesConfrontoRegione)
	sort.Strings(tempChoices)

	return tempChoices
}

// Sorts national fields selected for comparison
func (b *bot) getSortedChoicesConfrontoNazione() []string {
	tempChoices := make([]string, len(b.choicesConfrontoNazione))
	copy(tempChoices, b.choicesConfrontoNazione)
	sort.Strings(tempChoices)

	return tempChoices
}
