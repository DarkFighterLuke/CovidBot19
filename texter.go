package main

import (
	"fmt"
	"github.com/DarkFighterLuke/covidgraphs"
	"github.com/NicoNex/echotron"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

// Handles "report" textual command
func (b *bot) textReport(update *echotron.Update) {
	usageMessage := "<b>Uso Corretto del Comando:</b>\n/reports <code>[file] nome_report</code>\nReport disponibili:{<code>" +
		strings.Join(reports, ", ") + "</code>}\nDigita /help per visualizzare il manuale."

	tokens := strings.Fields(update.Message.Text)
	tokens = tokens[1:]

	var fieldNames []string
	var flagFile bool = false

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
		"/nazione <code>nome_dei_campi</code>\nper ottenere un confronto tra campi a tua scelta\n" +
		"Dati nazione disponibili:\n{<code>" + strings.Join(natregAttributes, ", ") + "</code>}\nDigita /help per visualizzare il manuale."

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

		b.SendPhoto(filename, setCaptionConfrontoNazione(len(nationData)-1, fieldNames), update.Message.Chat.ID, echotron.PARSE_HTML)
	}
	b.DeleteMessage(update.Message.Chat.ID, update.Message.ID)
}

// Handles "regione" textual command
func (b *bot) textRegion(update *echotron.Update) {
	dirPath := workingDirectory + imageFolder
	usageMessage := "<b>Uso Corretto del Comando:\n</b>/regione <code>nome_regione andamento</code>\nper ottenere l'andamento della regione scelta\n" +
		"/regione <code>nome_regione nome_dei_campi</code>\nper ottenere un confronto tra campi a tua scelta sulla desiderata\n" +
		"Dati regione disponibili:\n{<code>" + strings.Join(natregAttributes, ", ") + "</code>}\nDigita /help per visualizzare il manuale."

	tokens := strings.Fields(update.Message.Text)
	tokens = tokens[1:]

	var fieldNames []string

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

		b.SendPhoto(filename, setCaptionConfrontoRegione(regionId, fieldNames), update.Message.Chat.ID, echotron.PARSE_HTML)
	}
	b.DeleteMessage(update.Message.Chat.ID, update.Message.ID)
}

// Handles "provincia" textual command
func (b *bot) textProvince(update *echotron.Update) {
	dirPath := workingDirectory + imageFolder
	usageMessage := "<b>Uso Corretto del Comando:\n</b>/provincia <code>nome_provincia totale_casi</code>" +
		"\nper ottenere informazioni sul totale dei casi della provincia scelta\n" +
		"/provincia <code>nome_provincia nuovi_positivi</code>\nper ottenere informazioni sui nuovi positivi della provincia scelta\nDigita /help per visualizzare il manuale."

	tokens := strings.Fields(update.Message.Text)
	tokens = tokens[1:]

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

		b.SendPhoto(filename, setCaptionConfrontoProvincia(provinceLastId, []string{tokens[1]}), update.Message.Chat.ID, echotron.PARSE_HTML)
	}
	b.DeleteMessage(update.Message.Chat.ID, update.Message.ID)
}
