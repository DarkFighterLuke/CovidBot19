package main

import (
	"fmt"
	"github.com/DarkFighterLuke/covidgraphs"
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
	buttons, err := b.buttonsConfrontoNazione()
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
	buttons, err := b.buttonsConfrontoRegione()
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
	homeButton, err := b.makeButtons([]string{"Torna alla Home"}, []string{"home"}, 1)
	if err != nil {
		log.Println(err)
		return
	}
	b.SendMessageWithKeyboard(setCaptionTopRegions(), cq.Message.Chat.ID, homeButton, echotron.PARSE_HTML)
	b.AnswerCallbackQuery(cq.ID, "Classifica zonesButtons", false)
}

func (b *bot) callbackClassificaProvince(cq *echotron.CallbackQuery) {
	//TODO: grafico a barre classifica
	homeButton, err := b.makeButtons([]string{"Torna alla Home"}, []string{"home"}, 1)
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
	buttons, err := b.makeButtons([]string{"Genera file", "Torna alla Home"}, []string{"genera_file", "home"}, 1)
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
			log.Println(err)
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
		buttons, err := b.buttonsCaseConfrontoRegione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "terapia intensiva regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "terapia intensiva")
		buttons, err := b.buttonsCaseConfrontoRegione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale ospedalizzati regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "totale ospedalizzati")
		buttons, err := b.buttonsCaseConfrontoRegione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "isolamento domiciliare regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "isolamento domiciliare")
		buttons, err := b.buttonsCaseConfrontoRegione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "attualmente positivi regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "attualmente positivi")
		buttons, err := b.buttonsCaseConfrontoRegione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "nuovi positivi regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "nuovi positivi")
		buttons, err := b.buttonsCaseConfrontoRegione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "dimessi guariti regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "dimessi guariti")
		buttons, err := b.buttonsCaseConfrontoRegione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "deceduti regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "deceduti")
		buttons, err := b.buttonsCaseConfrontoRegione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale casi regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "totale casi")
		buttons, err := b.buttonsCaseConfrontoRegione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "tamponi regione":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "tamponi")
		buttons, err := b.buttonsCaseConfrontoRegione()
		if err != nil {
			return err
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
		buttons, err := b.buttonsCaseConfrontoNazione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "terapia intensiva nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "terapia intensiva")
		buttons, err := b.buttonsCaseConfrontoNazione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale ospedalizzati nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "totale ospedalizzati")
		buttons, err := b.buttonsCaseConfrontoNazione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "isolamento domiciliare nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "isolamento domiciliare")
		buttons, err := b.buttonsCaseConfrontoNazione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "attualmente positivi nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "attualmente positivi")
		buttons, err := b.buttonsCaseConfrontoNazione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "nuovi positivi nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "nuovi positivi")
		buttons, err := b.buttonsCaseConfrontoNazione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "dimessi guariti nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "dimessi guariti")
		buttons, err := b.buttonsCaseConfrontoNazione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "deceduti nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "deceduti")
		buttons, err := b.buttonsCaseConfrontoNazione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale casi nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "totale casi")
		buttons, err := b.buttonsCaseConfrontoNazione()
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
	case "tamponi nazione":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "tamponi")
		buttons, err := b.buttonsCaseConfrontoNazione()
		if err != nil {
			return err
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

// Handles in group callbacks for selected nation fields to compare
func (b *bot) caseInGroupNationAttr(cq *echotron.CallbackQuery) error {
	switch strings.ToLower(cq.Data) {
	case "andamento nazione groups":
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		b.sendAndamentoNazionale(cq.Message)
		break
	case "previous nazione groups":
		id := b.lastGroupAttrIndex
	redoPrevious:
		if id <= 0 {
			id = 9
		} else {
			id--
		}

		buttons, err := b.buttonsConfrontoNazioneGroups(id)
		if err != nil {
			if err.Error() == "already chose" {
				goto redoPrevious
			}
			log.Println(err)
			return err
		}

		b.lastGroupAttrIndex = id
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break
	case "next nazione groups":
		id := b.lastGroupAttrIndex
	redoNext:
		if id >= 9 {

			id = 0
		} else {
			id++
		}

		buttons, err := b.buttonsConfrontoNazioneGroups(id)
		if err != nil {
			if err.Error() == "already chose" {
				goto redoNext
			}
			log.Println(err)
			return err
		}

		b.lastGroupAttrIndex = id
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break

	case "annulla nazione groups":
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		break

	case "ricoverati con sintomi nazione groups":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "ricoverati con sintomi")
		buttons, err := b.buttonsConfrontoNazioneGroups(1)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 1
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "terapia intensiva nazione groups":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "terapia intensiva")
		buttons, err := b.buttonsConfrontoNazioneGroups(2)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 2
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale ospedalizzati nazione groups":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "totale ospedalizzati")
		buttons, err := b.buttonsConfrontoNazioneGroups(3)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 3
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "isolamento domiciliare nazione groups":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "isolamento domiciliare")
		buttons, err := b.buttonsConfrontoNazioneGroups(4)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 4
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "attualmente positivi nazione groups":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "attualmente positivi")
		buttons, err := b.buttonsConfrontoNazioneGroups(5)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 5
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "nuovi positivi nazione groups":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "nuovi positivi")
		buttons, err := b.buttonsConfrontoNazioneGroups(6)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 6
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "dimessi guariti nazione groups":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "dimessi guariti")
		buttons, err := b.buttonsConfrontoNazioneGroups(7)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 7
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "deceduti nazione groups":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "deceduti")
		buttons, err := b.buttonsConfrontoNazioneGroups(8)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 8
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale casi nazione groups":
		b.choicesConfrontoNazione = append(b.choicesConfrontoNazione, "totale casi")
		buttons, err := b.buttonsConfrontoNazioneGroups(9)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 9
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "tamponi nazione groups":
		b.choicesConfrontoNazione = append(b.choicesConfrontoRegione, "tamponi")
		buttons, err := b.buttonsConfrontoNazioneGroups(10)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 10
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "fatto nazione groups":
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		if len(b.choicesConfrontoNazione) == 0 {
			b.sendAndamentoNazionale(cq.Message)
		} else {
			b.sendConfrontoDatiNazione(cq)
		}
		b.choicesConfrontoRegione = make([]string, 0)
		b.lastGroupRegionIndex = 0
		b.lastZoneIndex = 0
		b.lastGroupAttrIndex = 0
		break
	default:
		return fmt.Errorf("not a confronto nazione case")
	}

	return nil
}

// Handles in group callbacks for selected region fields to compare
func (b *bot) caseInGroupRegionAttr(cq *echotron.CallbackQuery) error {
	switch strings.ToLower(cq.Data) {
	case "andamento region attr groups":
		regionIndex, err := covidgraphs.FindLastOccurrenceRegion(&regionsData, "denominazione_regione", b.lastRegion)
		if err != nil {
			return err
		}
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		b.sendAndamentoRegionale(cq.Message, regionIndex)
		break
	case "previous region attr groups":
		id := b.lastGroupAttrIndex
	redoPrevious:
		if id <= 0 {
			id = 10
		} else {
			id--
		}

		buttons, err := b.buttonsConfrontoRegioneGroups(id)
		if err != nil {
			if err.Error() == "already chose" {
				goto redoPrevious
			}
			log.Println(err)
			return err
		}

		b.lastGroupAttrIndex = id
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break
	case "next region attr groups":
		id := b.lastGroupAttrIndex
	redoNext:
		if id >= 10 {

			id = 0
		} else {
			id++
		}

		buttons, err := b.buttonsConfrontoRegioneGroups(id)
		if err != nil {
			if err.Error() == "already chose" {
				goto redoNext
			}
			log.Println(err)
			return err
		}

		b.lastGroupAttrIndex = id
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break

	case "annulla region attr groups":
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		break

	case "ricoverati con sintomi region attr groups":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "ricoverati con sintomi")
		buttons, err := b.buttonsConfrontoRegioneGroups(1)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 1
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "terapia intensiva region attr groups":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "terapia intensiva")
		buttons, err := b.buttonsConfrontoRegioneGroups(2)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 2
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale ospedalizzati region attr groups":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "totale ospedalizzati")
		buttons, err := b.buttonsConfrontoRegioneGroups(3)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 3
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "isolamento domiciliare region attr groups":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "isolamento domiciliare")
		buttons, err := b.buttonsConfrontoRegioneGroups(4)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 4
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "attualmente positivi region attr groups":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "attualmente positivi")
		buttons, err := b.buttonsConfrontoRegioneGroups(5)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 5
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "nuovi positivi region attr groups":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "nuovi positivi")
		buttons, err := b.buttonsConfrontoRegioneGroups(6)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 6
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "dimessi guariti region attr groups":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "dimessi guariti")
		buttons, err := b.buttonsConfrontoRegioneGroups(7)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 7
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "deceduti region attr groups":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "deceduti")
		buttons, err := b.buttonsConfrontoRegioneGroups(8)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 8
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "totale casi region attr groups":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "totale casi")
		buttons, err := b.buttonsConfrontoRegioneGroups(9)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 9
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "tamponi region attr groups":
		b.choicesConfrontoRegione = append(b.choicesConfrontoRegione, "tamponi")
		buttons, err := b.buttonsConfrontoRegioneGroups(10)
		if err != nil {
			return err
		}

		b.lastGroupAttrIndex = 10
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "Aggiunto al confronto", false)
		break
	case "fatto region attr groups":
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		if len(b.choicesConfrontoRegione) == 0 {
			regionIndex, err := covidgraphs.FindLastOccurrenceRegion(&regionsData, "denominazione_regione", b.lastRegion)
			if err != nil {
				return err
			}
			b.sendAndamentoRegionale(cq.Message, regionIndex)
		} else {
			b.sendConfrontoDatiRegione(cq)
		}
		b.choicesConfrontoRegione = make([]string, 0)
		b.lastGroupRegionIndex = 0
		b.lastZoneIndex = 0
		b.lastGroupAttrIndex = 0
		b.lastRegion = ""
		break
	default:
		return fmt.Errorf("not a confronto nazione groups case")
	}

	return nil
}

func (b *bot) caseInGroupZone(cq *echotron.CallbackQuery) error {
	switch strings.ToLower(cq.Data) {
	case "previous zone groups":
		id := b.lastGroupAttrIndex
	redoPrevious:
		if id <= 0 {
			id = 2
		} else {
			id--
		}

		buttons, err := b.buttonsZonesGroups(id)
		if err != nil {
			if err.Error() == "already chose" {
				goto redoPrevious
			}
			log.Println(err)
			return err
		}

		b.lastGroupAttrIndex = id
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break
	case "next zone groups":
		id := b.lastGroupAttrIndex
	redoNext:
		if id >= 2 {

			id = 0
		} else {
			id++
		}

		buttons, err := b.buttonsZonesGroups(id)
		if err != nil {
			if err.Error() == "already chose" {
				goto redoNext
			}
			log.Println(err)
			return err
		}

		b.lastGroupAttrIndex = id
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break

	case "annulla zone groups":
		b.lastGroupAttrIndex = 0
		b.lastGroupRegionIndex = 0
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		break
	case "sud groups":
		buttons, err, lastRegionName, _ := b.buttonsRegionsGroups(0, 0)
		if err != nil {
			return err
		}

		b.lastZoneIndex = 0
		b.lastGroupAttrIndex = 0
		b.lastGroupRegionIndex = 0
		b.lastRegion = lastRegionName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break
	case "centro groups":
		buttons, err, lastRegionName, _ := b.buttonsRegionsGroups(1, 0)
		if err != nil {
			return err
		}

		b.lastZoneIndex = 1
		b.lastGroupAttrIndex = 0
		b.lastGroupRegionIndex = 0
		b.lastRegion = lastRegionName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break
	case "nord groups":
		buttons, err, lastRegionName, _ := b.buttonsRegionsGroups(2, 0)
		if err != nil {
			return err
		}

		b.lastZoneIndex = 2
		b.lastGroupAttrIndex = 0
		b.lastGroupRegionIndex = 0
		b.lastRegion = lastRegionName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break

	default:
		return fmt.Errorf("not a zone groups case")
	}

	return nil
}

func (b *bot) caseInGroupRegion(cq *echotron.CallbackQuery) error {
	switch strings.ToLower(cq.Data) {
	case "previous region groups":
		id := b.lastGroupRegionIndex
		id--

		buttons, err, lastRegionName, lastIndex := b.buttonsRegionsGroups(b.lastZoneIndex, id)
		if err != nil {
			return err
		}

		b.lastGroupRegionIndex = lastIndex
		b.lastRegion = lastRegionName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break
	case "next region groups":
		id := b.lastGroupRegionIndex
		id++

		buttons, err, lastRegionName, lastIndex := b.buttonsRegionsGroups(b.lastZoneIndex, id)
		if err != nil {
			return err
		}

		b.lastGroupRegionIndex = lastIndex
		b.lastRegion = lastRegionName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break

	case "annulla region groups":
		b.lastRegion = ""
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		break
	}
	_, err := covidgraphs.FindLastOccurrenceRegion(&regionsData, "denominazione_regione", strings.Replace(cq.Data, " groups", "", -1))
	if err == nil {
		buttons, err := b.buttonsConfrontoRegioneGroups(0)
		if err != nil {
			return err
		}

		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
	}

	return nil
}

func (b *bot) caseInGroupZoneP(cq *echotron.CallbackQuery) error {
	switch strings.ToLower(cq.Data) {
	case "previous zone p groups":
		id := b.lastGroupAttrIndex
	redoPrevious:
		if id <= 0 {
			id = 2
		} else {
			id--
		}

		buttons, err := b.buttonsZonesGroupsP(id)
		if err != nil {
			if err.Error() == "already chose" {
				goto redoPrevious
			}
			log.Println(err)
			return err
		}

		b.lastGroupAttrIndex = id
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break
	case "next zone p groups":
		id := b.lastGroupAttrIndex
	redoNext:
		if id >= 2 {

			id = 0
		} else {
			id++
		}

		buttons, err := b.buttonsZonesGroupsP(id)
		if err != nil {
			if err.Error() == "already chose" {
				goto redoNext
			}
			log.Println(err)
			return err
		}

		b.lastGroupAttrIndex = id
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break

	case "annulla zone p groups":
		b.lastGroupAttrIndex = 0
		b.lastGroupRegionIndex = 0
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		break
	case "sud p groups":
		buttons, err, lastRegionName, _ := b.buttonsRegionsGroupsP(0, 0)
		if err != nil {
			return err
		}

		b.lastZoneIndex = 0
		b.lastGroupAttrIndex = 0
		b.lastGroupRegionIndex = 0
		b.lastRegion = lastRegionName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break
	case "centro p groups":
		buttons, err, lastRegionName, _ := b.buttonsRegionsGroupsP(1, 0)
		if err != nil {
			return err
		}

		b.lastZoneIndex = 1
		b.lastGroupAttrIndex = 0
		b.lastGroupRegionIndex = 0
		b.lastRegion = lastRegionName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break
	case "nord p groups":
		buttons, err, lastRegionName, _ := b.buttonsRegionsGroupsP(2, 0)
		if err != nil {
			return err
		}

		b.lastZoneIndex = 2
		b.lastGroupAttrIndex = 0
		b.lastGroupRegionIndex = 0
		b.lastRegion = lastRegionName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break

	default:
		return fmt.Errorf("not a zone groups case")
	}

	return nil
}

func (b *bot) caseInGroupRegionP(cq *echotron.CallbackQuery) error {
	switch strings.ToLower(cq.Data) {
	case "previous region p groups":
		id := b.lastGroupRegionIndex
		id--

		buttons, err, lastRegionName, lastIndex := b.buttonsRegionsGroupsP(b.lastZoneIndex, id)
		if err != nil {
			return err
		}

		b.lastGroupRegionIndex = lastIndex
		b.lastRegion = lastRegionName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break
	case "next region p groups":
		id := b.lastGroupRegionIndex
		id++

		buttons, err, lastRegionName, lastIndex := b.buttonsRegionsGroupsP(b.lastZoneIndex, id)
		if err != nil {
			return err
		}

		b.lastGroupRegionIndex = lastIndex
		b.lastRegion = lastRegionName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break

	case "annulla region p groups":
		b.lastRegion = ""
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		break
	}
	_, err := covidgraphs.FindLastOccurrenceRegion(&regionsData, "denominazione_regione", strings.Replace(cq.Data, " p groups", "", -1))
	if err == nil {
		buttons, err, _, _ := b.buttonsProvincesGroup(0, b.lastRegion)
		if err != nil {
			return err
		}

		b.lastGroupProvinceIndex = 0
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
	}

	return err
}

func (b *bot) caseInGroupProvince(cq *echotron.CallbackQuery) error {
	switch strings.ToLower(cq.Data) {
	case "previous province groups":
		id := b.lastGroupProvinceIndex
		id--

		buttons, err, lastProvinceName, lastIndex := b.buttonsProvincesGroup(id, b.lastRegion)
		if err != nil {
			return err
		}

		b.lastGroupProvinceIndex = lastIndex
		b.lastProvince = lastProvinceName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break
	case "next province groups":
		id := b.lastGroupProvinceIndex
		id++

		buttons, err, lastProvinceName, lastIndex := b.buttonsProvincesGroup(id, b.lastRegion)
		if err != nil {
			return err
		}

		b.lastGroupProvinceIndex = lastIndex
		b.lastProvince = lastProvinceName
		b.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.ID, buttons)
		b.AnswerCallbackQuery(cq.ID, "", false)
		break

	case "annulla province groups":
		b.lastRegion = ""
		b.DeleteMessage(cq.Message.Chat.ID, cq.Message.ID)
		break
	}
	provinceId, err := covidgraphs.FindLastOccurrenceProvince(&provincesData, "denominazione_provincia", strings.Replace(cq.Data, " province", "", -1))
	if err == nil {
		b.sendAndamentoProvinciale(cq, provinceId)
	}

	return err
}
