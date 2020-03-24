package main

import (
	"encoding/json"
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"image/color"
	"io/ioutil"
	"net/http"
)

type Nation interface {
	ricoveratiConSintomiToPoints()
	terapiaIntensivaToPoints()
	totaleOspedalizzatiToPoints()
	isolamentoDomiciliareToPoints()
	totalePositiviToPoints()
	nuoviPositiviToPoints()
	totaleGuaritiToPoints()
	totaleDecedutiToPoints()
	totaleContagiToPoints()
	totaleCasiToPoints()
	totaleTamponiToPoints()
}

type Region interface {
	ricoveratiConSintomiToPoints()
	terapiaIntensivaToPoints()
	totaleOspedalizzatiToPoints()
	isolamentoDomiciliareToPoints()
	totalePositiviToPoints()
	nuoviPositiviToPoints()
	totaleGuaritiToPoints()
	totaleDecedutiToPoints()
	totaleContagiToPoints()
	totaleCasiToPoints()
	totaleTamponiToPoints()
}

type Province interface {
	totaleCasiToPoints()
}

type nationData struct {
	Data                        string `json:"data"`
	Stato                       string `json:"stato"`
	Ricoverati_con_sintomi      uint64 `json:"ricoverati_con_sintomi"`
	Terapia_intensiva           uint64 `json:"terapia_intensiva"`
	Totale_ospedalizzati        uint64 `json:"totale_ospedalizzati"`
	Isolamento_domiciliare      uint64 `json:"isolamento_domiciliare"`
	Totale_attualmente_positivi uint64 `json:"totale_attualmente_positivi"`
	Nuovi_attualmente_positivi  uint64 `json:"nuovi_attualmente_positivi"`
	Dimessi_guariti             uint64 `json:"dimessi_guariti"`
	Deceduti                    uint64 `json:"deceduti"`
	Totale_casi                 uint64 `json:"totale_casi"`
	Tamponi                     uint64 `json:"tamponi"`
}

type regionData struct {
	Data                        string  `json:"data"`
	Stato                       string  `json:"stato"`
	Codice_regione              uint64  `json:"codice_regione"`
	Denominazione_regione       string  `json:"denominazione_regione"`
	Lat                         float64 `json:"lat"`
	Long                        float64 `json:"long"`
	Ricoverati_con_sintomi      uint64  `json:"ricoverati_con_sintomi"`
	Terapia_intensiva           uint64  `json:"terapia_intensiva"`
	Totale_ospedalizzati        uint64  `json:"totale_ospedalizzati"`
	Isolamento_domiciliare      uint64  `json:"isolamento_domiciliare"`
	Totale_attualmente_positivi uint64  `json:"totale_attualmente_positivi"`
	Nuovi_attualmente_positivi  uint64  `json:"nuovi_attualmente_positivi"`
	Dimessi_guariti             uint64  `json:"dimessi_guariti"`
	Deceduti                    uint64  `json:"deceduti"`
	Totale_casi                 uint64  `json:"totale_casi"`
	Tamponi                     uint64  `json:"tamponi"`
}

type provinceData struct {
	Data                    string  `json:"data"`
	Stato                   string  `json:"stato"`
	Codice_regione          uint64  `json:"codice_regione"`
	Denominazione_regione   string  `json:"denominazione_regione"`
	Codice_provincia        uint64  `json:"codice_provincia"`
	Denominazione_provincia string  `json:"denominazione_provincia"`
	Sigla_provincia         string  `json:"sigla_provincia"`
	Lat                     float64 `json:"lat"`
	Long                    float64 `json:"long"`
	Totale_casi             uint64  `json:"totale_casi"`
}

type plotData struct {
	points       []plotter.XYs
	title        string
	xAxisName    string
	yAxisName    string
	linesColors  []color.RGBA
	pointsShape  []interface{}
	pointsColors []color.RGBA
	legendNames  []string
}

func getNazione() error {
	var response []nationData
	resp, err := http.Get("https://raw.githubusercontent.com/pcm-dpc/COVID-19/master/dati-json/dpc-covid19-ita-andamento-nazionale.json")
	if err != nil {
		return fmt.Errorf("error receiving data: %v", err)
	} else {
		var bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error in received body: %v", err)
		} else {
			err := json.Unmarshal(bodyBytes, &response)
			defer resp.Body.Close()
			if err != nil {
				return fmt.Errorf("error in json unmarshal: %v", err)
			}
		}
	}

	return nil
}

func getProvince() error {
	var response []provinceData
	resp, err := http.Get("https://raw.githubusercontent.com/pcm-dpc/COVID-19/master/dati-json/dpc-covid19-ita-province.json")
	if err != nil {
		return fmt.Errorf("error receiving data: %v", err)
	} else {
		var bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error in received body: %v", err)
		} else {
			err := json.Unmarshal(bodyBytes, &response)
			defer resp.Body.Close()
			if err != nil {
				return fmt.Errorf("error in json unmarshal: %v", err)
			}
		}
	}

	return nil
}

func getRegioni() error {
	var response []regionData
	resp, err := http.Get("https://raw.githubusercontent.com/pcm-dpc/COVID-19/master/dati-json/dpc-covid19-ita-regioni.json")
	if err != nil {
		return fmt.Errorf("error receiving data: %v", err)
	} else {
		var bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error in received body: %v", err)
		} else {
			err := json.Unmarshal(bodyBytes, &response)
			defer resp.Body.Close()
			if err != nil {
				return fmt.Errorf("error in json unmarshal: %v", err)
			}
		}
	}

	return nil
}

func dataScatterPlot(data plotData) error {
	p, err := plot.New()
	if err != nil {
		return fmt.Errorf("errore: %v", err)
	}
	p.Title.Text = data.title
	p.X.Label.Text = data.xAxisName
	p.Y.Label.Text = data.yAxisName

	p.Add(plotter.NewGrid())
	for i, set := range data.points {
		line, points, err := plotter.NewLinePoints(set)
		if err != nil {
			return fmt.Errorf("an error occurred while adding points to plot: %v", err)
		}
		line.Color = data.linesColors[i]

		var pointsShape draw.GlyphDrawer
		switch data.pointsShape[i].(type) {
		case draw.CircleGlyph:
			pointsShape = data.pointsShape[i].(draw.CircleGlyph)
		case draw.BoxGlyph:
			pointsShape = data.pointsShape[i].(draw.BoxGlyph)
		case draw.CrossGlyph:
			pointsShape = data.pointsShape[i].(draw.CrossGlyph)
		case draw.PlusGlyph:
			pointsShape = data.pointsShape[i].(draw.PlusGlyph)
		case draw.PyramidGlyph:
			pointsShape = data.pointsShape[i].(draw.PyramidGlyph)
		case draw.RingGlyph:
			pointsShape = data.pointsShape[i].(draw.RingGlyph)
		case draw.SquareGlyph:
			pointsShape = data.pointsShape[i].(draw.SquareGlyph)
		case draw.TriangleGlyph:
			pointsShape = data.pointsShape[i].(draw.TriangleGlyph)
		default:
			return fmt.Errorf("bad point shape format")
		}
		points.Shape = pointsShape
		points.Color = data.pointsColors[i]

		p.Add(line, points)
	}

	if err := p.Save(8*vg.Inch, 8*vg.Inch, "points.png"); err != nil {
		return fmt.Errorf("error saving the plot: %v", err)
	}
	return nil
}
