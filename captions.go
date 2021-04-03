package main

import (
	"covidgraphs"
	"log"
	"strconv"
	"time"
)

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
