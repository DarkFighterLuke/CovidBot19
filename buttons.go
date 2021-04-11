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

	keys := make([]echotron.InlineKbdRow, 0)
	for i, v := range buttonsText {
		buttons := make([]echotron.InlineButton, 0)
		for j := 0; j < layoutCols; j++ {
			if j > len(buttonsText)-i {
				break
			} else {
				buttons = append(buttons, b.InlineKbdBtn(buttonsText[i], "", callbacksData[i]))
			}
			i++
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

func (b *bot) buttonsCaseConfrontoNazione() ([]byte, error) {
	buttonsNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}
	buttonsCallback := make([]string, 0)

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
		return nil, err
	}

	return buttons, nil
}

func (b *bot) buttonsCaseConfrontoRegione() ([]byte, error) {
	buttonsNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}
	buttonsCallback := make([]string, 0)

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
		return nil, err
	}

	return buttons, nil
}

func (b *bot) buttonsConfrontoNazioneGroups(attributeIndex int) ([]byte, error) {
	attributeNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}

	extendedAttributeNames := []string{"Andamento"}
	extendedAttributeNames = append(extendedAttributeNames, attributeNames...)
	extendedAttributeCallbacks := make([]string, 0)
	for _, v := range extendedAttributeNames {
		extendedAttributeCallbacks = append(extendedAttributeCallbacks, strings.ToLower(v)+" nazione groups")
	}
	if attributeIndex >= len(extendedAttributeNames) || attributeIndex < 0 {
		return nil, fmt.Errorf("attributeIndex out of range")
	}

	if b.isStringFoundInNationChoices(extendedAttributeNames[attributeIndex]) {
		return nil, fmt.Errorf("already chose")
	}
	buttonNames := []string{"«", extendedAttributeNames[attributeIndex], "»", "❌", "✅"}
	buttonCallbacks := []string{"previous nazione groups", extendedAttributeCallbacks[attributeIndex], "next nazione groups", "annulla nazione groups", "fatto nazione groups"}
	buttons, err := b.makeButtons(buttonNames, buttonCallbacks, 3)
	if err != nil {
		return nil, err
	}

	return buttons, nil
}

func (b *bot) buttonsZonesGroups(zoneIndex int) ([]byte, error) {
	zoneNames := []string{"Sud", "Centro", "Nord"}
	zoneCallbacks := []string{"sud groups", "centro groups", "nord groups"}
	if zoneIndex >= len(zoneNames) || zoneIndex < 0 {
		return nil, fmt.Errorf("attributeIndex out of range")
	}

	buttonNames := []string{"«", zoneNames[zoneIndex], "»", "❌"}
	buttonCallbacks := []string{"previous zone groups", zoneCallbacks[zoneIndex], "next zone groups", "annulla zone groups"}
	buttons, err := b.makeButtons(buttonNames, buttonCallbacks, 3)
	if err != nil {
		return nil, err
	}

	return buttons, nil
}

func (b *bot) buttonsRegionsGroups(zoneIndex, regionIndex int) ([]byte, error, string, int) {
	if zoneIndex < 0 || zoneIndex > 2 {
		return nil, fmt.Errorf("zoneIndex out of range"), "", -1
	}

	var regionNames []string
	if zoneIndex == 0 {
		regionNames = covidgraphs.GetSudRegionsNamesList()
	} else if zoneIndex == 1 {
		regionNames = covidgraphs.GetCentroRegionsNamesList()
	} else if zoneIndex == 2 {
		regionNames = covidgraphs.GetNordRegionsNamesList()
	}

	regionCallbacks := make([]string, 0)
	for _, v := range regionNames {
		regionCallbacks = append(regionCallbacks, strings.ToLower(v)+" groups")
	}

	if regionIndex >= len(regionNames) {
		regionIndex = 0
	} else if regionIndex < 0 {
		regionIndex = len(regionNames) - 1
	}

	buttonNames := []string{"«", regionNames[regionIndex], "»", "❌"}
	buttonCallbacks := []string{"previous region groups", regionCallbacks[regionIndex], "next region groups", "annulla region groups"}
	buttons, err := b.makeButtons(buttonNames, buttonCallbacks, 3)
	if err != nil {
		return nil, err, "", -1
	}

	return buttons, nil, strings.ToLower(regionNames[regionIndex]), regionIndex
}

func (b *bot) buttonsConfrontoRegioneGroups(attributeIndex int) ([]byte, error) {
	attributeNames := []string{"Ricoverati con sintomi", "Terapia intensiva", "Totale ospedalizzati", "Isolamento domiciliare", "Attualmente positivi", "Nuovi positivi", "Dimessi guariti", "Deceduti", "Totale casi", "Tamponi"}
	extendedAttributeNames := []string{"Andamento"}
	extendedAttributeNames = append(extendedAttributeNames, attributeNames...)

	extendedAttributeCallback := make([]string, 0)
	for _, v := range extendedAttributeNames {
		extendedAttributeCallback = append(extendedAttributeCallback, strings.ToLower(v)+" region attr groups")
	}
	if attributeIndex >= len(extendedAttributeNames) || attributeIndex < 0 {
		return nil, fmt.Errorf("attributeIndex out of range")
	}

	if b.isStringFoundInNationChoices(extendedAttributeNames[attributeIndex]) {
		return nil, fmt.Errorf("already chose")
	}
	buttonNames := []string{"«", extendedAttributeNames[attributeIndex], "»", "❌", "✅"}
	buttonCallbacks := []string{"previous region attr groups", extendedAttributeCallback[attributeIndex], "next region attr groups", "annulla region attr groups", "fatto region attr groups"}
	buttons, err := b.makeButtons(buttonNames, buttonCallbacks, 3)
	if err != nil {
		return nil, err
	}

	return buttons, nil
}

func (b *bot) buttonsZonesGroupsP(zoneIndex int) ([]byte, error) {
	zoneNames := []string{"Sud", "Centro", "Nord"}
	zoneCallbacks := []string{"sud p groups", "centro p groups", "nord p groups"}
	if zoneIndex >= len(zoneNames) || zoneIndex < 0 {
		return nil, fmt.Errorf("attributeIndex out of range")
	}

	buttonNames := []string{"«", zoneNames[zoneIndex], "»", "❌"}
	buttonCallbacks := []string{"previous zone p groups", zoneCallbacks[zoneIndex], "next zone p groups", "annulla zone p groups"}
	buttons, err := b.makeButtons(buttonNames, buttonCallbacks, 3)
	if err != nil {
		return nil, err
	}

	return buttons, nil
}

func (b *bot) buttonsRegionsGroupsP(zoneIndex, regionIndex int) ([]byte, error, string, int) {
	if zoneIndex < 0 || zoneIndex > 2 {
		return nil, fmt.Errorf("zoneIndex out of range"), "", -1
	}

	var regionNames []string
	if zoneIndex == 0 {
		regionNames = covidgraphs.GetSudRegionsNamesList()
	} else if zoneIndex == 1 {
		regionNames = covidgraphs.GetCentroRegionsNamesList()
	} else if zoneIndex == 2 {
		regionNames = covidgraphs.GetNordRegionsNamesList()
	}

	regionCallbacks := make([]string, 0)
	for _, v := range regionNames {
		regionCallbacks = append(regionCallbacks, strings.ToLower(v)+" p groups")
	}

	if regionIndex >= len(regionNames) {
		regionIndex = 0
	} else if regionIndex < 0 {
		regionIndex = len(regionNames) - 1
	}

	buttonNames := []string{"«", regionNames[regionIndex], "»", "❌"}
	buttonCallbacks := []string{"previous region p groups", regionCallbacks[regionIndex], "next region p groups", "annulla region p groups"}
	buttons, err := b.makeButtons(buttonNames, buttonCallbacks, 3)
	if err != nil {
		return nil, err, "", -1
	}

	return buttons, nil, strings.ToLower(regionNames[regionIndex]), regionIndex
}

func (b *bot) buttonsProvincesGroup(provinceIndex int, regionName string) ([]byte, error, string, int) {
	provinces := covidgraphs.GetLastProvincesByRegionName(&provincesData, strings.ToLower(regionName))
	provinceNames := make([]string, 0)
	for _, v := range *provinces {
		provinceNames = append(provinceNames, v.Denominazione_provincia)
	}
	provinceCallbacks := make([]string, 0)
	for _, v := range *provinces {
		provinceCallbacks = append(provinceCallbacks, strings.ToLower(v.Denominazione_provincia)+" province")
	}

	if provinceIndex >= len(provinceNames) {
		provinceIndex = 0
	} else if provinceIndex < 0 {
		provinceIndex = len(provinceNames) - 1
	}

	buttonNames := []string{"«", provinceNames[provinceIndex], "»", "❌"}
	buttonCallbacks := []string{"previous province groups", provinceCallbacks[provinceIndex], "next province groups", "annulla province groups"}

	buttons, err := b.makeButtons(buttonNames, buttonCallbacks, 3)
	if err != nil {
		return nil, err, "", -1
	}

	return buttons, nil, strings.ToLower(provinceNames[provinceIndex]), provinceIndex
}
