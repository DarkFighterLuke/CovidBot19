package main

import (
	"fmt"
	"github.com/DarkFighterLuke/covidgraphs"
	"github.com/NicoNex/echotron"
	"log"
	"strings"
)

// Creates buttons sets
func (b *bot) makeButtons(buttonsText []string, callbacksData []string, layoutCols int) ([]byte, error) {
	if len(buttonsText) != len(callbacksData) || layoutCols <= 0 {
		return nil, fmt.Errorf("different text and data length")
	}

	buttons := make([]echotron.InlineButton, 0)
	keys := make([]echotron.InlineKbdRow, 0)
	for i, v := range buttonsText {
		for j := 0; j < layoutCols; j++ {
			if j > len(buttonsText)-i {
				break
			} else {
				buttons = append(buttons, b.InlineKbdBtn(v, "", callbacksData[i]))
			}
		}
		keys = append(keys, buttons)
	}

	inlineKMarkup := b.InlineKbdMarkup(keys...)
	return inlineKMarkup, nil
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
func (b *bot) nordRegionsButtons() ([]byte, error) {
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
func (b *bot) centroRegionsButtons() ([]byte, error) {
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
func (b *bot) sudRegionsButtons() ([]byte, error) {
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

func (b *bot) buttonsConfrontoNazione() ([]byte, error) {
	buttonsNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}
	buttonsCallback := make([]string, 0)
	for _, v := range buttonsNames {
		buttonsCallback = append(buttonsCallback, strings.ToLower(v)+" nazione")
	}
	buttonsNames = append(buttonsNames, "Annulla ❌")
	buttonsCallback = append(buttonsCallback, "annulla")
	buttons, err := b.makeButtons(buttonsNames, buttonsCallback, 2)
	if err != nil {
		return nil, err
	}

	return buttons, nil
}

func (b *bot) buttonsConfrontoRegione() ([]byte, error) {
	buttonsNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}
	buttonsCallback := make([]string, 0)
	for _, v := range buttonsNames {
		buttonsCallback = append(buttonsCallback, strings.ToLower(v)+" regione")
	}
	buttonsNames = append(buttonsNames, "Annulla ❌")
	buttonsCallback = append(buttonsCallback, "annulla")
	buttons, err := b.makeButtons(buttonsNames, buttonsCallback, 2)
	if err != nil {
		return nil, err
	}

	return buttons, nil
}
