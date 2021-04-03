package main

import (
	"covidgraphs"
	"fmt"
	"github.com/NicoNex/echotron"
	"log"
	"os"
	"strings"
	"time"
)

func (b *bot) callbackNuoviCasiNazione(cq *echotron.CallbackQuery) {
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

	buttons, err := b.makeButtons([]string{"Torna alla Home"}, []string{"sendHome"}, 1)
	if err != nil {
		log.Println(err)
		return
	}

	b.SendPhotoWithKeyboard(filename, setCaptionConfrontoNazione(len(nationData)-1, []string{"nuovi_positivi"}), cq.Message.Chat.ID, buttons, echotron.PARSE_HTML)
	b.AnswerCallbackQuery(cq.ID, "Nuovi casi", false)
	b.lastButton = "zonesButtons"
	b.lastRegion = ""
	b.lastProvince = ""
}

func (b *bot) callbackNuoviCasiRegione(cq *echotron.CallbackQuery) {
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

	buttons, err := b.makeButtons([]string{"Torna alla Regione", "Torna alla Home"}, []string{b.lastRegion, "sendHome"}, 1)
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
}

func (b *bot) callbackStoricoNazione(cq *echotron.CallbackQuery) {
	b.lastButton = cq.Data
	b.lastRegion = ""
	b.lastProvince = ""
}

func (b *bot) callbackZonesButtons(cq *echotron.CallbackQuery) {
	buttons, err := b.zonesButtons()
	if err != nil {
		log.Println(err)
	}
	b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
	b.AnswerCallbackQuery(cq.ID, "Zone", false)
	b.lastButton = cq.Data
	b.lastRegion = ""
	b.lastProvince = ""
}

func (b *bot) callbackConfrontoDatiNazione(cq *echotron.CallbackQuery) {
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
	b.EditMessageTextWithKeyboard(cq.Message.Chat.ID, cq.Message.ID, "Seleziona i campi che vuoi mettere a confronto:", buttons, echotron.PARSE_HTML)
	b.AnswerCallbackQuery(cq.ID, "Crea confronto dati nazione", false)
	b.lastButton = "zonesButtons"
	b.lastRegion = ""
	b.lastProvince = ""
}

func (b *bot) callbackConfrontoDatiRegione(cq *echotron.CallbackQuery) {
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
	b.EditMessageTextWithKeyboard(cq.Message.Chat.ID, cq.Message.ID, "Seleziona i campi che vuoi mettere a confronto:", buttons)
	b.AnswerCallbackQuery(cq.ID, "Crea confronto dati regione", false)
	b.lastButton = "province"
	b.lastProvince = ""
}

func (b *bot) callbackClassificaRegioni(cq *echotron.CallbackQuery) {
	//TODO: grafico a barre classifica
	homeButton, err := b.makeButtons([]string{"Torna alla Home"}, []string{"sendHome"}, 1)
	if err != nil {
		log.Println(err)
		return
	}
	b.SendMessageWithKeyboard(setCaptionTopRegions(), cq.Message.Chat.ID, homeButton, echotron.PARSE_HTML)
	b.AnswerCallbackQuery(cq.ID, "Classifica zonesButtons", false)
}

func (b *bot) callbackClassificaProvince(cq *echotron.CallbackQuery) {
	//TODO: grafico a barre classifica
	homeButton, err := b.makeButtons([]string{"Torna alla Home"}, []string{"sendHome"}, 1)
	if err != nil {
		log.Println(err)
		return
	}
	b.SendMessageWithKeyboard(setCaptionTopProvinces(), cq.Message.Chat.ID, homeButton, echotron.PARSE_HTML)
	b.AnswerCallbackQuery(cq.ID, "Classifica province", false)
}

func (b *bot) callbackNord(cq *echotron.CallbackQuery) {
	buttons, err := b.nordRegionsButtons()
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
}

func (b *bot) callbackCentro(cq *echotron.CallbackQuery) {
	buttons, err := b.centroRegionsButtons()
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
}

func (b *bot) callbackSud(cq *echotron.CallbackQuery) {
	buttons, err := b.sudRegionsButtons()
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
}

func (b *bot) callbackProvince(cq *echotron.CallbackQuery) {
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
}

func (b *bot) callbackHome(cq *echotron.CallbackQuery) {
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
}

func (b *bot) callbackReports(cq *echotron.CallbackQuery) {
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
}

func (b *bot) callbackReportGenerale(cq *echotron.CallbackQuery) {
	buttons, err := b.makeButtons([]string{"Genera file", "Torna alla Home"}, []string{"genera_file", "sendHome"}, 1)
	if err != nil {
		log.Println("Errore", err)
		return
	}

	msg := setCaptionAndamentoNazionale() + "\n\n\n" + setCaptionTopRegions() + "\n" + setCaptionTopProvinces()
	b.SendMessageWithKeyboard(msg, cq.Message.Chat.ID, buttons, echotron.PARSE_HTML)
	b.AnswerCallbackQuery(cq.ID, "Report generale", false)
	b.lastButton = "report generale"
}

func (b *bot) callbackGeneraFile(cq *echotron.CallbackQuery) {
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
