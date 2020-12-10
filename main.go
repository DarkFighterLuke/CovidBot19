package main

import (
	"covidgraphs"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"github.com/yanzay/tbot"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	nTopRegions      = 10
	botDataDirectory = "/CovidBot/"
	imageFolder      = "/plots/"
	logsFolder       = "/logs/"
)

var workingDirectory string

// User runtime data struct
type application struct {
	client                  *tbot.Client // Client instance representing the user using the bot
	dailyUpdate             bool         // This is a spoiler for a future feature
	lastButton              string       // Callback of the last pressed button
	lastRegion              string       // Callback of the last pressed button in case it is a region name
	lastProvince            string       // Callback of the last pressed button in case it is a province name
	choicesConfrontoNazione []string     // National fields selected for comparison
	choicesConfrontoRegione []string     // Regional fields selected for comparison
}

var nationData []covidgraphs.NationData      // National data array
var regionsData []covidgraphs.RegionData     // Regional data array
var provincesData []covidgraphs.ProvinceData // Provincial data array
var datiNote []covidgraphs.NoteData          // Notes array

var natregAttributes = []string{"ricoverati_con_sintomi", "terapia_intensiva", "totale_ospedalizzati",
	"isolamento_domiciliare", "totale_positivi", "nuovi_positivi", "dimessi_guariti", "deceduti",
	"totale_casi", "tamponi"} // National and regional fields names

var reports = []string{"generale"} // Types of reports avvailable

var mutex = &sync.Mutex{} // Mutex used when updating data from the pcm-dpc repo

// Creates bot data folders if they don't exist
func initFolders() {
	currentPath, _ := os.Getwd()
	workingDirectory = currentPath + botDataDirectory
	os.MkdirAll(workingDirectory, 0755)
	os.MkdirAll(workingDirectory+imageFolder, 0755)
	os.MkdirAll(workingDirectory+logsFolder, 0755)
}

func main() {
	log.SetOutput(os.Stdout)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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
	bot := tbot.New(os.Getenv("CovidBot"))
	//bot := tbot.New(os.Getenv("CovidBot"), tbot.WithWebhook("https://covid19bot.tk/bot", ":443"))

	app := &application{}
	app.client = bot.Client()

	// Handling "start" command
	bot.HandleMessage("/start", func(m *tbot.Message) {
		if m.Chat.Type == "private" {
			buttons, err := makeButtons([]string{"Credits 🌟", "Vai ai Dati 📊"}, []string{"credits", "home"}, 1)
			if err != nil {
				log.Println(err)
			}
			_, err = app.client.SendMessage(m.Chat.ID, "Benvenuto <b>"+m.Chat.FirstName+"</b>!\nQuesto bot mette a disposizione i dati "+
				"dell'epidemia di Coronavirus in Italia con grafici e numeri.\n\n"+
				"Puoi seguire i pulsanti per ottenere comodamente le informazioni che desideri\n"+
				"\n<b><i>oppure</i></b>\n"+
				"\nSe ti piace digitare puoi usare i seguenti comandi:\n"+
				"/nazione <code>andamento</code>\nper ottenere l'andamento della nazione\n"+
				"/nazione <code>nome_dei_campi</code>\nper ottenere un confronto tra campi a tua scelta\n\n"+
				"/regione <code>nome_regione andamento</code>\nper ottenere l'andamento della regione scelta\n"+
				"/regione <code>nome_regione nome_dei_campi</code>\nper ottenere un confronto tra campi a tua scelta sulla desiderata\n\n"+
				"/provincia <code>nome_provincia totale_casi</code>\nper ottenere informazioni sul totale dei casi della provincia scelta\n"+
				"/provincia <code>nome_provincia nuovi_positivi</code>\nper ottenere informazioni sui nuovi positivi della provincia scelta\n\n"+
				"/reports <code>[file] nome_report</code>\n\n\n"+
				"Dati nazione disponibili:\n{<code>"+strings.Join(natregAttributes, ",")+"</code>}\n\n"+
				"Dati regione disponibili:\n{<code>"+strings.Join(natregAttributes, ",")+"</code>}\n\n"+
				"Report disponibili:\n{<code>"+strings.Join(reports, ",")+"</code>}\n\n"+
				"Questi comandi possono sempre tornarti utili! Prova ad <b>aggiungere il bot in un gruppo</b> per tenere informate le tue cerchie.\n"+
				"\nCominciamo!", tbot.OptParseModeHTML, tbot.OptInlineKeyboardMarkup(buttons))
			if err != nil {
				log.Println("errore nell'invio del messaggio di benvenuto")
			}
		}
	})

	// Handling textual commands
	bot.HandleMessage("/nazione *", app.textNation)
	bot.HandleMessage("/regione *", app.textRegion)
	bot.HandleMessage("/provincia *", app.textProvince)
	bot.HandleMessage("/reports *", app.textReport)
	bot.HandleMessage("/home", app.home)
	bot.HandleMessage("/credits", app.credits)

	// Handling buttons callbacks
	bot.HandleCallback(app.callbackHandler)

	// Starting bot instance
	err := bot.Start()
	if err != nil {
		log.Fatal(err)
	}
}

// Handles "home" command
func (app *application) home(m *tbot.Message) {
	log.Println("[HOME] Nation: " + fmt.Sprint(nationData) + "\nRegions: " + fmt.Sprint(regionsData) + "\nProvinces: " + fmt.Sprint(provincesData) + "\t" + m.Chat.FirstName + "\n")
	if m.Chat.Type == "private" {
		app.sendAndamentoNazionale(m)
		buttons, err := app.mainMenuButtons()
		if err != nil {
			log.Println(err)
		}
		app.client.SendMessage(m.Chat.ID, "Scegli un'opzione", tbot.OptInlineKeyboardMarkup(buttons))
	}
}

// Handles "report" textual command
func (app *application) textReport(m *tbot.Message) {
	message := strings.Replace(m.Text, "/reports ", "", 1)
	var fieldNames []string
	var flagFile bool = false
	tokens := strings.Split(message, " ")
	if len(tokens) < 1 {
		if m.Chat.Type == "private" {
			app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:\n"+
				"/reports <code>[file] nome_report</code>\nDigita /start per visualizzare il manuale.", tbot.OptParseModeHTML)
		}
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
			if m.Chat.Type == "private" {
				app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:</b>\n/reports <code>[file] nome_report</code>", tbot.OptParseModeMarkdown)
			}
			return
		}
	}
	for _, v := range fieldNames {
		if v == "generale" {
			msg := setCaptionAndamentoNazionale() + "\n\n\n" + setCaptionTopRegions() + "\n" + setCaptionTopProvinces()
			app.client.SendMessage(m.Chat.ID, msg, tbot.OptParseModeHTML)
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

				app.client.DeleteMessage(m.Chat.ID, m.MessageID)
				app.client.SendDocumentFile(m.Chat.ID, filename)
				err = os.Remove(filename)
				if err != nil {
					log.Println("can't delete file " + filename)
				}
			}
		}
	}
}

// Handles "nazione" textual command
func (app *application) textNation(m *tbot.Message) {
	message := strings.Replace(m.Text, "/nazione ", "", 1)
	var fieldNames []string
	tokens := strings.Split(message, " ")
	if len(tokens) < 1 {
		if m.Chat.Type == "private" {
			app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:\n</b>/nazione <code>andamento</code>\nper ottenere l'andamento della nazione\n"+
				"/nazione <code>nome_dei_campi</code>\nper ottenere un confronto tra campi a tua scelta\nDigita /start per visualizzare il manuale.", tbot.OptParseModeHTML)
		}
		return
	}

	sort.Strings(natregAttributes)
	if tokens[0] == "andamento" {
		app.sendAndamentoNazionale(m)
	} else {
		for i := 0; i < len(tokens); i++ {
			if res := sort.SearchStrings(natregAttributes, tokens[i]); res < len(natregAttributes) {
				fieldNames = append(fieldNames, tokens[i])
			} else {
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
			if m.Chat.Type == "private" {
				app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:\n</b>/nazione <code>andamento</code>\nper ottenere l'andamento della nazione\n"+
					"/nazione <code>nome_dei_campi</code>\nper ottenere un confronto tra campi a tua scelta\nDigita /start per visualizzare il manuale.", tbot.OptParseModeHTML)
			}
			return
		}
		app.client.DeleteMessage(m.Chat.ID, m.MessageID)
		app.client.SendPhotoFile(m.Chat.ID, filename, tbot.OptCaption(setCaptionConfrontoNazione(len(nationData)-1, fieldNames)), tbot.OptParseModeHTML)
	}
}

// Handles "regione" textual command
func (app *application) textRegion(m *tbot.Message) {
	dirPath := workingDirectory + imageFolder

	message := strings.Replace(m.Text, "/regione ", "", 1)
	var fieldNames []string
	tokens := strings.Split(message, " ")

	if len(tokens) < 2 {
		if m.Chat.Type == "private" {
			app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:\n</b>/regione <code>nome_regione andamento</code>\nper ottenere l'andamento della regione scelta\n"+
				"/regione <code>nome_regione nome_dei_campi</code>\nper ottenere un confronto tra campi a tua scelta sulla desiderata\nDigita /start per visualizzare il manuale.", tbot.OptParseModeHTML)
		}
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
		app.sendAndamentoRegionale(m, regionId)
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
			if m.Chat.Type == "private" {
				app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:\n</b>/regione <code>nome_regione andamento</code>\nper ottenere l'andamento della regione scelta\n"+
					"/regione <code>nome_regione nome_dei_campi</code>\nper ottenere un confronto tra campi a tua scelta sulla desiderata\nDigita /start per visualizzare il manuale.", tbot.OptParseModeHTML)
			}
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
			if m.Chat.Type == "private" {
				app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:\n</b>/regione <code>nome_regione andamento</code>\nper ottenere l'andamento della regione scelta\n"+
					"/regione <code>nome_regione nome_dei_campi</code>\nper ottenere un confronto tra campi a tua scelta sulla desiderata\nDigita /start per visualizzare il manuale.", tbot.OptParseModeHTML)
			}
			return
		}

		app.client.DeleteMessage(m.Chat.ID, m.MessageID)
		app.client.SendPhotoFile(m.Chat.ID, filename, tbot.OptCaption(setCaptionConfrontoRegione(regionId, fieldNames)), tbot.OptParseModeHTML)
	}
}

// Handles "provincia" textual command
func (app *application) textProvince(m *tbot.Message) {
	dirPath := workingDirectory + imageFolder

	message := strings.Replace(m.Text, "/provincia ", "", 1)
	tokens := strings.Split(message, " ")
	if len(tokens) < 2 {
		if m.Chat.Type == "private" {
			app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:\n</b>/provincia <code>nome_provincia totale_casi</code>"+
				"\nper ottenere informazioni sul totale dei casi della provincia scelta\n"+
				"/provincia <code>nome_provincia nuovi_positivi</code>\nper ottenere informazioni sui nuovi positivi della provincia scelta\nDigita /start per visualizzare il manuale.", tbot.OptParseModeHTML)
		}
		return
	}
	if len(tokens) > 2 {
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
			if m.Chat.Type == "private" {
				app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:\n</b>/provincia <code>nome_provincia totale_casi</code>"+
					"\nper ottenere informazioni sul totale dei casi della provincia scelta\n"+
					"/provincia <code>nome_provincia nuovi_positivi</code>\nper ottenere informazioni sui nuovi positivi della provincia scelta\nDigita /start per visualizzare il manuale.", tbot.OptParseModeMarkdown)
			}
			return
		}

		provinceLastId, err := covidgraphs.FindLastOccurrenceProvince(&provincesData, "denominazione_provincia", tokens[0])
		if err != nil {
			log.Println(err)
			if m.Chat.Type == "private" {
				app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:\n</b>/provincia <code>nome_provincia totale_casi</code>"+
					"\nper ottenere informazioni sul totale dei casi della provincia scelta\n"+
					"/provincia <code>nome_provincia nuovi_positivi</code>\nper ottenere informazioni sui nuovi positivi della provincia scelta\nDigita /start per visualizzare il manuale.", tbot.OptParseModeMarkdown)
			}
			return
		}

		app.client.SendPhotoFile(m.Chat.ID, filename, tbot.OptCaption(setCaptionConfrontoProvincia(provinceLastId, []string{tokens[1]})), tbot.OptParseModeHTML)
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
			if m.Chat.Type == "private" {
				app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:\n</b>/provincia <code>nome_provincia totale_casi</code>"+
					"\nper ottenere informazioni sul totale dei casi della provincia scelta\n"+
					"/provincia <code>nome_provincia nuovi_positivi</code>\nper ottenere informazioni sui nuovi positivi della provincia scelta\nDigita /start per visualizzare il manuale.", tbot.OptParseModeMarkdown)
			}
			return
		}

		provinceLastId, err := covidgraphs.FindLastOccurrenceProvince(&provincesData, "denominazione_provincia", tokens[0])
		if err != nil {
			log.Println(err)
			app.client.SendMessage(m.Chat.ID, "<b>Uso Corretto del Comando:\n</b>/provincia <code>nome_provincia totale_casi</code>"+
				"\nper ottenere informazioni sul totale dei casi della provincia scelta\n"+
				"/provincia <code>nome_provincia nuovi_positivi</code>\nper ottenere informazioni sui nuovi positivi della provincia scelta\nDigita /start per visualizzare il manuale.", tbot.OptParseModeMarkdown)
			return
		}

		app.client.DeleteMessage(m.Chat.ID, m.MessageID)
		app.client.SendPhotoFile(m.Chat.ID, filename, tbot.OptCaption(setCaptionConfrontoProvincia(provinceLastId, []string{tokens[1]})), tbot.OptParseModeHTML)
	}
}

// Updates data from pcm-dpc repository
func updateData(nazione *[]covidgraphs.NationData, regioni *[]covidgraphs.RegionData, province *[]covidgraphs.ProvinceData, note *[]covidgraphs.NoteData) func() {
	return func() {
		mutex.Lock()

		covidgraphs.DeleteAllPlots(imageFolder)

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
func (app *application) sendAndamentoNazionale(m *tbot.Message) {
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
			app.client.SendMessage(m.Chat.ID, "Impossibile reperire il grafico al momento.\nRiprova più tardi.")
			return
		}
	}

	app.client.SendPhotoFile(m.Chat.ID, filename, tbot.OptCaption(setCaptionAndamentoNazionale()), tbot.OptParseModeHTML)
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

// Recognize received buttons callbacks
func (app *application) callbackHandler(cq *tbot.CallbackQuery) {
	writeOperation(cq)

	switch strings.ToLower(cq.Data) {
	case "credits":
		app.credits(cq.Message)
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Crediti"))
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

		buttons, err := makeButtons([]string{"Torna alla Home"}, []string{"home"}, 1)
		if err != nil {
			log.Println(err)
			return
		}
		app.client.SendPhotoFile(cq.Message.Chat.ID, filename, tbot.OptCaption(setCaptionConfrontoNazione(len(nationData)-1, []string{"nuovi_positivi"})), tbot.OptInlineKeyboardMarkup(buttons), tbot.OptParseModeHTML)
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Nuovi casi"))
		app.lastButton = "zonesButtons"
		app.lastRegion = ""
		app.lastProvince = ""
		break
	case "nuovi casi regione":
		regionId, err := covidgraphs.FindFirstOccurrenceRegion(&regionsData, "denominazione_regione", app.lastRegion)
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

		buttons, err := makeButtons([]string{"Torna alla Regione", "Torna alla Home"}, []string{app.lastRegion, "home"}, 1)
		if err != nil {
			log.Println(err)
			return
		}
		regionLastId, err := covidgraphs.FindLastOccurrenceRegion(&regionsData, "denominazione_regione", app.lastRegion)
		if err != nil {
			log.Println(err)
			return
		}
		app.client.SendPhotoFile(cq.Message.Chat.ID, filename, tbot.OptCaption(setCaptionConfrontoRegione(regionLastId, []string{"nuovi_positivi"})), tbot.OptInlineKeyboardMarkup(buttons), tbot.OptParseModeHTML)
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Nuovi casi"))
		app.lastButton = "province"
		app.lastProvince = ""
		break
	case "storico nazione":
		app.lastButton = cq.Data
		app.lastRegion = ""
		app.lastProvince = ""
		break
	case "zonesbuttons":
		buttons, err := app.zonesButtons()
		if err != nil {
			log.Println(err)
		}
		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Zone"))
		app.lastButton = cq.Data
		app.lastRegion = ""
		app.lastProvince = ""
		break
	case "vai a regione":
		app.client.SendMessage(cq.Message.Chat.ID, "Scrivi il nome di una regione di cui ottenere i dati:")
		break
	case "vai a provincia":
		app.client.SendMessage(cq.Message.Chat.ID, "Scrivi il nome di una provincia di cui ottenere i dati:")
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
		buttons, err := makeButtons(buttonsNames, buttonsCallback, 2)
		if err != nil {
			log.Println(err)
			return
		}
		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptText("Seleziona i campi che vuoi mettere a confronto:"), tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Crea confronto dati nazione"))
		app.lastButton = "zonesButtons"
		app.lastRegion = ""
		app.lastProvince = ""
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
		buttons, err := makeButtons(buttonsNames, buttonsCallback, 2)
		if err != nil {
			log.Println(err)
			return
		}
		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptText("Seleziona i campi che vuoi mettere a confronto:"), tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Crea confronto dati regione"))
		app.lastButton = "province"
		app.lastProvince = ""
		break
	case "classifica regioni":
		//TODO: grafico a barre classifica
		homeButton, err := makeButtons([]string{"Torna alla Home"}, []string{"home"}, 1)
		if err != nil {
			log.Println(err)
			return
		}
		app.client.SendMessage(cq.Message.Chat.ID, setCaptionTopRegions(), tbot.OptInlineKeyboardMarkup(homeButton), tbot.OptParseModeHTML)
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Classifica zonesButtons"))
		break
	case "classifica province":
		//TODO: grafico a barre classifica
		homeButton, err := makeButtons([]string{"Torna alla Home"}, []string{"home"}, 1)
		if err != nil {
			log.Println(err)
			return
		}
		app.client.SendMessage(cq.Message.Chat.ID, setCaptionTopProvinces(), tbot.OptInlineKeyboardMarkup(homeButton), tbot.OptParseModeHTML)
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Classifica province"))
		break

	case "nord":
		buttons, err := app.nordRegions()
		if err != nil {
			log.Println(err)
		}
		_, err = app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		if err != nil {
			log.Println(err)
		}
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Nord"))
		app.lastButton = cq.Data
		app.lastRegion = ""
		app.lastProvince = ""
		break
	case "centro":
		buttons, err := app.centroRegions()
		if err != nil {
			log.Println(err)
		}
		_, err = app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		if err != nil {
			log.Println(err)
		}
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Centro"))
		app.lastButton = cq.Data
		app.lastRegion = ""
		app.lastProvince = ""
		break
	case "sud":
		buttons, err := app.sudRegions()
		if err != nil {
			log.Println(err)
		}
		_, err = app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		if err != nil {
			log.Println(err)
		}
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Sud"))
		app.lastButton = cq.Data
		app.lastRegion = ""
		app.lastProvince = ""
		break

	case "province":
		provinces := covidgraphs.GetLastProvincesByRegionName(&provincesData, app.lastRegion)
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

		buttons, err := makeButtons(provincesNames, provincesCallback, 2)
		if err != nil {
			log.Println(err)
			app.client.SendMessage(cq.Message.Chat.ID, "Impossibile reperire il grafico al momento.\nRiprova più tardi.")
			return
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Province "+app.lastButton))
		app.lastButton = "province"

	case "home":
		app.sendAndamentoNazionale(cq.Message)
		buttons, err := app.mainMenuButtons()
		if err != nil {
			log.Println(err)
		}
		app.client.DeleteMessage(cq.Message.Chat.ID, cq.Message.MessageID)
		app.client.SendMessage(cq.Message.Chat.ID, "Scegli un'opzione", tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Home"))
		app.lastButton = cq.Data
		app.lastRegion = ""
		app.lastProvince = ""
		break
	case "annulla":
		app.back(cq)
		break

	case "reports":
		buttonsNames := []string{"Report generale"}
		buttonsCallback := []string{"report generale"}
		buttonsNames = append(buttonsNames, "Annulla ❌")
		buttonsCallback = append(buttonsCallback, "annulla")
		buttons, err := makeButtons(buttonsNames, buttonsCallback, 1)
		if err != nil {
			log.Println(err)
			return
		}
		app.client.SendMessage(cq.Message.Chat.ID, "Seleziona un tipo di report:", tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Reports"))
		app.lastButton = "reports"
		break
	case "report generale":
		buttons, err := makeButtons([]string{"Genera file", "Torna alla Home"}, []string{"genera file", "home"}, 1)
		if err != nil {
			log.Println(err)
			return
		}
		msg := setCaptionAndamentoNazionale() + "\n\n\n" + setCaptionTopRegions() + "\n" + setCaptionTopProvinces()
		app.client.SendMessage(cq.Message.Chat.ID, msg, tbot.OptInlineKeyboardMarkup(buttons), tbot.OptParseModeHTML)
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Report generale"))
		app.lastButton = "report generale"
	case "genera file":
		msg := setCaptionAndamentoNazionale() + "\n\n\n" + setCaptionTopRegions() + "\n" + setCaptionTopProvinces()
		switch app.lastButton {
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
				app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Si è verificato un errore"))
				return
			}
			err = f.Close()
			if err != nil {
				log.Println(err)
				app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Si è verificato un errore"))
				return
			}
			app.client.SendDocumentFile(cq.Message.Chat.ID, filename)
			app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Report generato"))
			err = os.Remove(filename)
			if err != nil {
				log.Println("can't delete file " + filename)
			}
		}
		break

	default:
		if _, err := covidgraphs.FindFirstOccurrenceRegion(&regionsData, "denominazione_regione", cq.Data); err == nil {
			app.caseRegion(cq)
		} else if _, err = covidgraphs.FindFirstOccurrenceProvince(&provincesData, "denominazione_provincia", cq.Data); err == nil {
			app.caseProvince(cq)
		} else if err = app.caseConfrontoRegione(cq); err == nil {
			break
		} else if err = app.caseConfrontoNazione(cq); err == nil {
			break
		} else {
			log.Println("dati callback incorretti")
		}
	}

}

// Creates buttons sets
func makeButtons(buttonsText []string, callbacksData []string, layout int) (*tbot.InlineKeyboardMarkup, error) {
	if layout != 1 && layout != 2 {
		return nil, fmt.Errorf("wrong layout")
	}
	if len(buttonsText) != len(callbacksData) {
		return nil, fmt.Errorf("different text and data length")
	}

	buttons := make([]tbot.InlineKeyboardButton, 0)
	for i, v := range buttonsText {
		buttons = append(buttons, tbot.InlineKeyboardButton{
			Text:         v,
			CallbackData: callbacksData[i],
		})
	}

	keys := make([][]tbot.InlineKeyboardButton, 0)
	switch layout {
	case 1:
		for i := 0; i < len(buttons); i++ {
			keys = append(keys, []tbot.InlineKeyboardButton{buttons[i]})
		}
		break
	case 2:
		for i := 0; i < len(buttons); i += 2 {
			if i+1 < len(buttons) {
				keys = append(keys, []tbot.InlineKeyboardButton{buttons[i], buttons[i+1]})
			} else {
				keys = append(keys, []tbot.InlineKeyboardButton{buttons[i]})
			}
		}
		break
	}

	return &tbot.InlineKeyboardMarkup{InlineKeyboard: keys}, nil
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
func (app *application) caseRegion(cq *tbot.CallbackQuery) {
	regionIndex, err := covidgraphs.FindLastOccurrenceRegion(&regionsData, "denominazione_regione", cq.Data)
	if err != nil {
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Si è verificato un errore"))
		return
	}
	app.client.DeleteMessage(cq.Message.Chat.ID, cq.Message.MessageID)
	app.sendAndamentoRegionale(cq.Message, regionIndex)
	buttons, err := app.provinceButtons()
	if err != nil {
		log.Println(err)
		return
	}
	app.client.SendMessage(cq.Message.Chat.ID, "Opzioni disponibili:", tbot.OptInlineKeyboardMarkup(buttons))
	app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Regione "+regionsData[regionIndex].Denominazione_regione))
	app.lastButton = cq.Data
	app.lastRegion = cq.Data
	app.lastProvince = ""
}

// Sends a region trend plot and text with related buttons
func (app *application) sendAndamentoRegionale(m *tbot.Message, regionIndex int) {
	firstRegionIndex, err := covidgraphs.FindFirstOccurrenceRegion(&regionsData, "codice_regione", regionsData[regionIndex].Codice_regione)
	if err != nil {
		app.client.SendMessage(m.Chat.ID, "Impossibile reperire il grafico al momento.\nRiprova più tardi.")
		return
	}

	dirPath := workingDirectory + imageFolder
	title := "Dati regione" + regionsData[firstRegionIndex].Denominazione_regione
	var filename string

	filename = dirPath + covidgraphs.FilenameCreator(title)
	if !covidgraphs.IsGraphExisting(filename) {
		err, filename = covidgraphs.VociRegione(&regionsData, []string{"totale_casi", "dimessi_guariti", "deceduti"}, 0, firstRegionIndex, title, filename)

		if err != nil {
			app.client.SendMessage(m.Chat.ID, "Impossibile reperire il grafico al momento.\nRiprova più tardi.")
			return
		}
	}

	app.client.SendPhotoFile(m.Chat.ID, filename, tbot.OptCaption(setCaptionRegion(regionIndex)), tbot.OptParseModeHTML)
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
func (app *application) caseProvince(cq *tbot.CallbackQuery) {
	provinceIndex, err := covidgraphs.FindLastOccurrenceProvince(&provincesData, "denominazione_provincia", cq.Data)
	if err != nil {
		log.Printf("province not found %v", err)
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Si è verificato un errore"))
		return
	}
	app.sendAndamentoProvinciale(cq, provinceIndex)
}

// Sends a province trend plot and text with related buttons
func (app *application) sendAndamentoProvinciale(cq *tbot.CallbackQuery, provinceIndex int) {
	dirPath := workingDirectory + imageFolder
	title := "Totale Contagi " + provincesData[provinceIndex].Denominazione_provincia
	var filename string
	var err error

	filename = dirPath + covidgraphs.FilenameCreator(title)
	if !covidgraphs.IsGraphExisting(filename) {
		provinceIndexes := covidgraphs.GetProvinceIndexesByName(&provincesData, provincesData[provinceIndex].Denominazione_provincia)
		err, filename = covidgraphs.TotalePositiviProvincia(&provincesData, provinceIndexes, title, filename)

		if err != nil {
			app.client.SendMessage(cq.Message.Chat.ID, "Impossibile reperire il grafico al momento.\nRiprova più tardi.")
			return
		}
	}

	buttonsNames := []string{"Torna alla regione", "Torna alla home"}
	callbackNames := []string{app.lastRegion, "home"}
	buttons, err := makeButtons(buttonsNames, callbackNames, 1)
	if err != nil {
		log.Println(err)
		return
	}

	app.client.SendPhotoFile(cq.Message.Chat.ID, filename, tbot.OptCaption(setCaptionProvince(provinceIndex)), tbot.OptParseModeHTML)
	app.client.SendMessage(cq.Message.Chat.ID, "Opzioni disponibili:", tbot.OptInlineKeyboardMarkup(buttons))
	app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Regione "+provincesData[provinceIndex].Denominazione_regione))
	app.lastButton = cq.Data
	app.lastProvince = cq.Data
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
func (app *application) back(cq *tbot.CallbackQuery) {
	writeOperation(cq)
	switch app.lastButton {
	case "zonesButtons":
		buttons, err := app.mainMenuButtons()
		if err != nil {
			log.Println(err)
			return
		}
		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Annulla"))
		app.choicesConfrontoRegione = make([]string, 0)
		app.choicesConfrontoNazione = make([]string, 0)
		break
	case "province":
		buttons, err := app.provinceButtons()
		if err != nil {
			log.Println(err)
			return
		}
		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Annulla"))
		app.lastButton = "province"
		app.choicesConfrontoRegione = make([]string, 0)
		app.choicesConfrontoNazione = make([]string, 0)
		break
	case "nord":
		buttons, err := app.zonesButtons()
		if err != nil {
			log.Println(err)
			return
		}
		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Annulla"))
		app.lastButton = "zonesButtons"
		break
	case "centro":
		buttons, err := app.zonesButtons()
		if err != nil {
			log.Println(err)
			return
		}
		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Annulla"))
		app.lastButton = "zonesButtons"
		break
	case "sud":
		buttons, err := app.zonesButtons()
		if err != nil {
			log.Println(err)
			return
		}
		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Annulla"))
		app.lastButton = "zonesButtons"
		break
	case "reports":
		app.client.DeleteMessage(cq.Message.Chat.ID, cq.Message.MessageID)
		app.lastButton = ""
		app.lastRegion = ""
		app.lastProvince = ""
	}
}

// Creates the main menu buttons set
func (app *application) mainMenuButtons() (*tbot.InlineKeyboardMarkup, error) {
	//buttonsNames := []string{"Storico 🕑", "Regioni", "Vai a regione ➡️", "Vai a provincia ➡️", "Crea confronto su dati nazione 📈", "Classifica regioni 🏅", "Classifica province 🏅"}
	buttonsNames := []string{"Nuovi casi 🆕", "Regioni", "Crea confronto su dati nazione 📈", "Classifica regioni 🏅", "Classifica province 🏅", "Reports 📃"}
	//callbackData := []string{"storico nazione", "zonesButtons", "vai a regione", "vai a provincia", "crea confronto su dati nazione", "classifica regioni", "classifica province"}
	callbackData := []string{"nuovi casi nazione", "zonesButtons", "confronto dati nazione", "classifica regioni", "classifica province", "reports"}
	buttons, err := makeButtons(buttonsNames, callbackData, 1)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	app.choicesConfrontoNazione = make([]string, 0)
	app.choicesConfrontoRegione = make([]string, 0)
	return buttons, nil
}

// Creates zones buttons set
func (app *application) zonesButtons() (*tbot.InlineKeyboardMarkup, error) {
	zones := []string{"Nord", "Centro", "Sud"}
	zonesCallback := make([]string, 0)
	for _, v := range zones {
		zonesCallback = append(zonesCallback, strings.ToLower(v))
	}
	zones = append(zones, "Annulla ❌")
	zonesCallback = append(zonesCallback, "annulla")
	buttons, err := makeButtons(zones, zonesCallback, 1)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return buttons, nil
}

// Creates provinces buttons set
func (app *application) provinceButtons() (*tbot.InlineKeyboardMarkup, error) {
	buttonsNames := []string{"Nuovi casi 🆕", "Province della regione", "Confronto dati regione 📈", "Torna alla home"}
	callbackNames := []string{"nuovi casi regione", "province", "confronto dati regione", "home"}
	buttons, err := makeButtons(buttonsNames, callbackNames, 1)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	app.choicesConfrontoRegione = make([]string, 0)
	return buttons, nil
}

// Creates northern regions buttons set
func (app *application) nordRegions() (*tbot.InlineKeyboardMarkup, error) {
	regions := covidgraphs.GetNordRegionsNamesList()
	regionsCallback := make([]string, 0)
	for _, v := range regions {
		regionsCallback = append(regionsCallback, strings.ToLower(v))
	}
	regions = append(regions, "Annulla ❌")
	regionsCallback = append(regionsCallback, "annulla")
	buttons, err := makeButtons(regions, regionsCallback, 2)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return buttons, nil
}

// Creates central regions buttons set
func (app *application) centroRegions() (*tbot.InlineKeyboardMarkup, error) {
	regions := covidgraphs.GetCentroRegionsNamesList()
	regionsCallback := make([]string, 0)
	for _, v := range regions {
		regionsCallback = append(regionsCallback, strings.ToLower(v))
	}
	regions = append(regions, "Annulla ❌")
	regionsCallback = append(regionsCallback, "annulla")
	buttons, err := makeButtons(regions, regionsCallback, 2)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return buttons, nil
}

// Creates southern regions buttons set
func (app *application) sudRegions() (*tbot.InlineKeyboardMarkup, error) {
	regions := covidgraphs.GetSudRegionsNamesList()
	regionsCallback := make([]string, 0)
	for _, v := range regions {
		regionsCallback = append(regionsCallback, strings.ToLower(v))
	}
	regions = append(regions, "Annulla ❌")
	regionsCallback = append(regionsCallback, "annulla")
	buttons, err := makeButtons(regions, regionsCallback, 2)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return buttons, nil
}

// Handles "Confronto dati regione" selected fields
func (app *application) caseConfrontoRegione(cq *tbot.CallbackQuery) error {
	buttonsNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}
	buttonsCallback := make([]string, 0)
	for _, v := range buttonsNames {
		buttonsCallback = append(buttonsCallback, strings.ToLower(v)+" regione")
	}

	switch cq.Data {
	case "ricoverati con sintomi regione":
		app.choicesConfrontoRegione = append(app.choicesConfrontoRegione, "ricoverati con sintomi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !app.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "terapia intensiva regione":
		app.choicesConfrontoRegione = append(app.choicesConfrontoRegione, "terapia intensiva")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !app.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "totale ospedalizzati regione":
		app.choicesConfrontoRegione = append(app.choicesConfrontoRegione, "totale ospedalizzati")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !app.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "isolamento domiciliare regione":
		app.choicesConfrontoRegione = append(app.choicesConfrontoRegione, "isolamento domiciliare")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !app.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "attualmente positivi regione":
		app.choicesConfrontoRegione = append(app.choicesConfrontoRegione, "attualmente positivi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !app.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "nuovi positivi regione":
		app.choicesConfrontoRegione = append(app.choicesConfrontoRegione, "nuovi positivi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !app.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "dimessi guariti regione":
		app.choicesConfrontoRegione = append(app.choicesConfrontoRegione, "dimessi guariti")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !app.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "deceduti regione":
		app.choicesConfrontoRegione = append(app.choicesConfrontoRegione, "deceduti")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !app.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "totale casi regione":
		app.choicesConfrontoRegione = append(app.choicesConfrontoRegione, "totale casi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !app.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "tamponi regione":
		app.choicesConfrontoRegione = append(app.choicesConfrontoRegione, "tamponi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInRegionChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " regione", "", -1)
			if !app.isStringFoundInRegionChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto regione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "fatto regione":
		app.client.DeleteMessage(cq.Message.Chat.ID, cq.Message.MessageID)
		app.sendConfrontoDatiRegione(cq)
		app.choicesConfrontoRegione = make([]string, 0)
		break
	default:
		return fmt.Errorf("not a confronto regioni case")
	}

	return nil
}

// Sends a plot with a caption containing a comparison with the selected regional fields
func (app *application) sendConfrontoDatiRegione(cq *tbot.CallbackQuery) {
	snakeCaseChoices := make([]string, 0)
	for _, v := range app.choicesConfrontoRegione {
		snakeCaseChoices = append(snakeCaseChoices, strings.Replace(v, " ", "_", -1))
	}
	app.choicesConfrontoRegione = snakeCaseChoices

	regionId, err := covidgraphs.FindFirstOccurrenceRegion(&regionsData, "denominazione_regione", app.lastRegion)
	if err != nil {
		log.Println(err)
		return
	}

	sortedChoices := app.getSortedChoicesConfrontoRegione()

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
		err, filename = covidgraphs.VociRegione(&regionsData, app.choicesConfrontoRegione, 0, regionId, title, filename)

		if err != nil {
			log.Println(err)
		}
	}

	buttonsNames := []string{"Torna alla regione", "Torna alla home"}
	callbackNames := []string{app.lastRegion, "home"}
	buttons, err := makeButtons(buttonsNames, callbackNames, 1)
	if err != nil {
		log.Println(err)
		return
	}
	regionLastId, err := covidgraphs.FindLastOccurrenceRegion(&regionsData, "denominazione_regione", regionsData[regionId].Denominazione_regione)
	if err != nil {
		log.Println(err)
		return
	}

	app.client.SendPhotoFile(cq.Message.Chat.ID, filename, tbot.OptCaption(setCaptionConfrontoRegione(regionLastId, app.choicesConfrontoRegione)), tbot.OptParseModeHTML)
	app.client.SendMessage(cq.Message.Chat.ID, "Opzioni disponibili:", tbot.OptInlineKeyboardMarkup(buttons))
	app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Confronto effettuato"))
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
func (app *application) isStringFoundInRegionChoices(str string) bool {
	for _, j := range app.choicesConfrontoRegione {
		if j == strings.ToLower(str) {
			return true
		}
	}
	return false
}

// Handles "Confronto dati nazione" selected fields
func (app *application) caseConfrontoNazione(cq *tbot.CallbackQuery) error {
	buttonsNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}
	buttonsCallback := make([]string, 0)
	for _, v := range buttonsNames {
		buttonsCallback = append(buttonsCallback, strings.ToLower(v)+" nazione")
	}

	switch cq.Data {
	case "ricoverati con sintomi nazione":
		app.choicesConfrontoNazione = append(app.choicesConfrontoNazione, "ricoverati con sintomi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !app.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "terapia intensiva nazione":
		app.choicesConfrontoNazione = append(app.choicesConfrontoNazione, "terapia intensiva")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !app.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "totale ospedalizzati nazione":
		app.choicesConfrontoNazione = append(app.choicesConfrontoNazione, "totale ospedalizzati")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !app.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "isolamento domiciliare nazione":
		app.choicesConfrontoNazione = append(app.choicesConfrontoNazione, "isolamento domiciliare")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !app.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "attualmente positivi nazione":
		app.choicesConfrontoNazione = append(app.choicesConfrontoNazione, "attualmente positivi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !app.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "nuovi positivi nazione":
		app.choicesConfrontoNazione = append(app.choicesConfrontoNazione, "nuovi positivi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !app.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "dimessi guariti nazione":
		app.choicesConfrontoNazione = append(app.choicesConfrontoNazione, "dimessi guariti")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !app.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "deceduti nazione":
		app.choicesConfrontoNazione = append(app.choicesConfrontoNazione, "deceduti")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !app.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "totale casi nazione":
		app.choicesConfrontoNazione = append(app.choicesConfrontoNazione, "totale casi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !app.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "tamponi nazione":
		app.choicesConfrontoNazione = append(app.choicesConfrontoNazione, "tamponi")
		newButtonsNames := make([]string, 0)
		newButtonsCallback := make([]string, 0)

		for _, v := range buttonsNames {

			if !app.isStringFoundInNationChoices(v) {
				newButtonsNames = append(newButtonsNames, v)
			}
		}
		for _, v := range buttonsCallback {
			strStripped := strings.Replace(v, " nazione", "", -1)
			if !app.isStringFoundInNationChoices(strStripped) {
				newButtonsCallback = append(newButtonsCallback, strings.ToLower(v))
			}
		}

		newButtonsNames = append(newButtonsNames, "Annulla ❌")
		newButtonsCallback = append(newButtonsCallback, "annulla")
		newButtonsNames = append(newButtonsNames, "Fatto ✅")
		newButtonsCallback = append(newButtonsCallback, "fatto nazione")
		buttons, err := makeButtons(newButtonsNames, newButtonsCallback, 2)
		if err != nil {
			log.Println(err)
		}

		app.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tbot.OptInlineKeyboardMarkup(buttons))
		app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Aggiunto al confronto"))
		break
	case "fatto nazione":
		app.client.DeleteMessage(cq.Message.Chat.ID, cq.Message.MessageID)
		app.sendConfrontoDatiNazione(cq)
		app.choicesConfrontoNazione = make([]string, 0)
		break
	default:
		return fmt.Errorf("not a confronto nazione case")
	}

	return nil
}

// Sends a plot with a caption containing a comparison with the selected national fields
func (app *application) sendConfrontoDatiNazione(cq *tbot.CallbackQuery) {
	snakeCaseChoices := make([]string, 0)
	for _, v := range app.choicesConfrontoNazione {
		snakeCaseChoices = append(snakeCaseChoices, strings.Replace(v, " ", "_", -1))
	}
	app.choicesConfrontoNazione = snakeCaseChoices

	sortedChoices := app.getSortedChoicesConfrontoNazione()

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
		err, filename = covidgraphs.VociNazione(&nationData, app.choicesConfrontoNazione, 0, title, filename)

		if err != nil {
			log.Println(err)
		}
	}

	buttonsNames := []string{"Torna alla home"}
	callbackNames := []string{"home"}
	buttons, err := makeButtons(buttonsNames, callbackNames, 1)
	if err != nil {
		log.Println(err)
		return
	}

	app.client.SendPhotoFile(cq.Message.Chat.ID, filename, tbot.OptCaption(setCaptionConfrontoNazione(len(nationData)-1, app.choicesConfrontoNazione)), tbot.OptParseModeHTML)
	app.client.SendMessage(cq.Message.Chat.ID, "Opzioni disponibili:", tbot.OptInlineKeyboardMarkup(buttons))
	app.client.AnswerCallbackQuery(cq.ID, tbot.OptText("Confronto effettuato"))
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
func (app *application) isStringFoundInNationChoices(str string) bool {
	for _, j := range app.choicesConfrontoNazione {
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
func writeOperation(cq *tbot.CallbackQuery) {
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
func (app *application) credits(m *tbot.Message) {
	buttons, err := makeButtons([]string{"Torna alla Home"}, []string{"home"}, 1)
	if err != nil {
		log.Println(err)
	}
	app.client.SendMessage(m.Chat.ID, "🤖 Bot creato da @GiovanniRanaTortello\n😺 GitHub: https://github.com/DarkFighterLuke\n"+
		"\n🌐 Proudly hosted on Raspberry Pi 3 powered by Arch Linux", tbot.OptInlineKeyboardMarkup(buttons))
}

// Sorts regional fields selected for comparison
func (app *application) getSortedChoicesConfrontoRegione() []string {
	tempChoices := make([]string, len(app.choicesConfrontoRegione))
	copy(tempChoices, app.choicesConfrontoRegione)
	sort.Strings(tempChoices)

	return tempChoices
}

// Sorts national fields selected for comparison
func (app *application) getSortedChoicesConfrontoNazione() []string {
	tempChoices := make([]string, len(app.choicesConfrontoNazione))
	copy(tempChoices, app.choicesConfrontoRegione)
	sort.Strings(tempChoices)

	return tempChoices
}
