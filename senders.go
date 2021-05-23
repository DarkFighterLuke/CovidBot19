package main

import (
	"fmt"
	"github.com/DarkFighterLuke/covidgraphs"
	"github.com/NicoNex/echotron"
	"log"
	"strings"
)

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

// Sends a region trend plot and text with related buttons
func (b *bot) sendAndamentoRegionale(message *echotron.Message, regionIndex int) {
	firstRegionIndex, err := covidgraphs.FindFirstOccurrenceRegion(&regionsData, "codice_regione", regionsData[regionIndex].Codice_regione)
	if err != nil {
		b.SendMessage("Impossibile reperire il grafico al momento.\nRiprova più tardi.", message.Chat.ID)
		return
	}

	dirPath := workingDirectory + imageFolder
	title := "Dati regione " + regionsData[firstRegionIndex].Denominazione_regione
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
	if cq.Message.Chat.Type == "private" {
		b.SendMessageWithKeyboard("Opzioni disponibili:", cq.Message.Chat.ID, buttons)
	}
	b.AnswerCallbackQuery(cq.ID, "Regione "+provincesData[provinceIndex].Denominazione_regione, false)
	b.lastButton = cq.Data
	b.lastProvince = cq.Data
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
	if cq.Message.Chat.Type == "private" {
		b.SendMessageWithKeyboard("Opzioni disponibili:", cq.Message.Chat.ID, buttons)
	}
	b.AnswerCallbackQuery(cq.ID, "Confronto effettuato", false)
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
	if cq.Message.Chat.Type == "private" {
		b.SendMessageWithKeyboard("Opzioni disponibili:", cq.Message.Chat.ID, buttons)
	}
	b.AnswerCallbackQuery(cq.ID, "Confronto effettuato", false)
}

func (b *bot) sendHelp(update *echotron.Update) {
	b.SendMessage(helpMsg, update.Message.Chat.ID, echotron.PARSE_HTML)
}

// Handles "start" command
func (b *bot) sendStart(update *echotron.Update) {
	if update.Message.Chat.Type == "private" {
		buttons, err := b.makeButtons([]string{"Credits 🌟", "Vai ai Dati 📊"}, []string{"credits", "home"}, 1)
		if err != nil {
			log.Println(err)
		}

		messageText :=
			`Benvenuto <b>%s</b>! Questo bot mette a disposizione i dati dell'epidemia di Coronavirus in Italia con grafici e numeri.
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

// Handles "home" command
func (b *bot) sendHome(update *echotron.Update) {
	if update.Message.Chat.Type == "private" {
		b.sendAndamentoNazionale(update.Message)
		buttons, err := b.mainMenuButtons()
		if err != nil {
			log.Println(err)
		}
		b.SendMessageWithKeyboard("Scegli un opzione", update.Message.Chat.ID, buttons)
	}
}

// Sends credits message
func (b *bot) sendCredits(chatId int64) {
	buttons, err := b.makeButtons([]string{"Torna alla Home"}, []string{"home"}, 1)
	if err != nil {
		log.Println(err)
	}
	b.SendMessageWithKeyboard("🤖 Bot creato da @GiovanniRanaTortello\n😺 GitHub: https://github.com/DarkFighterLuke\n"+
		"\n🌐 Proudly hosted on Raspberry Pi 3", chatId, buttons, echotron.PARSE_HTML)
}
