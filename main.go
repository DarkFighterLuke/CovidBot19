package main

import (
	"fmt"
	"github.com/DarkFighterLuke/covidgraphs"
	"github.com/DarkFighterLuke/gitUpdateChecker/v2"
	"github.com/NicoNex/echotron"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	nTopRegions      = 10
	botDataDirectory = "CovidBot"
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
	lastGroupRegionIndex    int
	lastGroupAttrIndex      int
	lastZoneIndex           int
	lastGroupProvinceIndex  int
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

var TOKEN = os.Getenv("CovidBot")

func newBot(chatId int64) echotron.Bot {
	return &bot{
		chatId: chatId,
		Api:    echotron.NewApi(TOKEN),
	}
}

func checkUpdate(nazione *[]covidgraphs.NationData, regioni *[]covidgraphs.RegionData, province *[]covidgraphs.ProvinceData, note *[]covidgraphs.NoteData, frequency time.Duration, stop chan bool) {
	log.Println("Starting update checker...")
	_ = gitUpdateChecker.SetRepoInfo("https://github.com/pcm-dpc/COVID-19.git", "master")
	ch, err := gitUpdateChecker.StartUpdateProcess(frequency)
	if err != nil {
		log.Println(err)
	}

	for {
		select {
		case u := <-ch:
			if u {
				log.Println("There is a new commit on pandemic data repository. Waiting 5 minutes to let update raw files...")
				time.Sleep(5 * time.Minute)
				log.Println("Retrieving data...")
				updateData(nazione, regioni, province, note)()
			}
		case s := <-stop:
			if s {
				log.Println("Stopping update checker...")
				return
			}
		}
	}
}

func main() {
	log.SetOutput(os.Stdout)
	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	initFolders()
	updateData(&nationData, &regionsData, &provincesData, &datiNote)()

	stop := make(chan bool)
	loc, _ := time.LoadLocation("Europe/Rome")
	now := time.Now()
	startHour := time.Date(now.Year(), now.Month(), now.Day(), 16, 00, 00, 00, loc)
	endHour := time.Date(now.Year(), now.Month(), now.Day(), 19, 00, 00, 00, loc)
	if time.Now().After(startHour) && time.Now().Before(endHour) {
		go checkUpdate(&nationData, &regionsData, &provincesData, &datiNote, 30*time.Second, stop)
	}

	// Planning cronjobs to update data from pcm-dpc repo
	var cronjob = cron.New()
	_, _ = cronjob.AddFunc("CRON_TZ=Europe/Rome 00 16 * * *", func() { checkUpdate(&nationData, &regionsData, &provincesData, &datiNote, 30*time.Second, stop) })
	_, _ = cronjob.AddFunc("CRON_TZ=Europe/Rome 00 19 * * *", func() { stop <- true })
	cronjob.Start()

	// Creating bot instance using webhook mode
	dsp := echotron.NewDispatcher(TOKEN, newBot)
	dsp.ListenWebhook("https://hiddenfile.ml:443/bot/CovidBot", 40987)
}

func (b *bot) Update(update *echotron.Update) {
	writeOperation(update, botDataDirectory+logsFolder)
	if update.Message != nil {
		keywords := strings.Split(update.Message.Text, " ")
		if keywords[0] == "/start" || keywords[0] == "/start"+botUsername {
			b.sendStart(update)
		} else if keywords[0] == "/help" || keywords[0] == "/help"+botUsername {
			b.sendHelp(update)
		} else if keywords[0] == "/home" || keywords[0] == "/home"+botUsername {
			b.sendHome(update)
		} else if keywords[0] == "/nazione" {
			b.textNation(update)
		} else if keywords[0] == "/nazione"+botUsername {
			b.inGroupTextNation(update.Message.Chat.ID)
		} else if keywords[0] == "/regione" {
			b.textRegion(update)
		} else if keywords[0] == "/regione"+botUsername {
			b.inGroupTextRegions(update.Message.Chat.ID)
		} else if keywords[0] == "/provincia" {
			b.textProvince(update)
		} else if keywords[0] == "/provincia"+botUsername {
			b.inGroupTextProvinces(update.Message.Chat.ID)
		} else if keywords[0] == "/reports" || keywords[0] == "/reports"+botUsername {
			b.textReport(update)
		} else if keywords[0] == "/credits" || keywords[0] == "/credits"+botUsername {
			b.sendCredits(update.Message.Chat.ID)
		}

	} else if update.CallbackQuery != nil {
		cq := update.CallbackQuery
		switch strings.ToLower(cq.Data) {
		case "credits":
			b.sendCredits(update.CallbackQuery.Message.Chat.ID)
			b.AnswerCallbackQuery(cq.ID, "Crediti", false)
		case "nuovi casi nazione":
			b.callbackNuoviCasiNazione(cq)
			break
		case "nuovi casi regione":
			b.callbackNuoviCasiRegione(cq)
			break
		case "storico nazione":
			b.callbackStoricoNazione(cq)
			break
		case "zonesbuttons":
			b.callbackZonesButtons(cq)
			break
		case "confronto dati nazione":
			b.callbackConfrontoDatiNazione(cq)
			break
		case "confronto dati regione":
			b.callbackConfrontoDatiRegione(cq)
			break
		case "classifica regioni":
			b.callbackClassificaRegioni(cq)
			break
		case "classifica province":
			b.callbackClassificaProvince(cq)
			break

		case "nord":
			b.callbackNord(cq)
			break
		case "centro":
			b.callbackCentro(cq)
			break
		case "sud":
			b.callbackSud(cq)
			break

		case "province":
			b.callbackProvince(cq)
			break

		case "home":
			b.callbackHome(cq)
			break
		case "annulla":
			b.back(cq)
			break

		case "reports":
			b.callbackReports(cq)
			break
		case "report generale":
			b.callbackReportGenerale(cq)
			break
		case "genera_file":
			b.callbackGeneraFile(cq)
			break

		default:
			//DON'T CHANGE THIS ORDER, OR CALLBACK HANDLING WILL FAIL!
			if _, err := covidgraphs.FindFirstOccurrenceRegion(&regionsData, "denominazione_regione", cq.Data); err == nil {
				b.caseRegion(cq)
			} else if _, err = covidgraphs.FindFirstOccurrenceProvince(&provincesData, "denominazione_provincia", cq.Data); err == nil {
				b.caseProvince(cq)
			} else if err = b.caseConfrontoRegione(cq); err == nil {
				break
			} else if err = b.caseConfrontoNazione(cq); err == nil {
				break
			} else if err = b.caseInGroupNationAttr(cq); err == nil {
				break
			} else if err = b.caseInGroupZoneP(cq); err == nil {
				break
			} else if err = b.caseInGroupRegionP(cq); err == nil {
				break
			} else if err = b.caseInGroupProvince(cq); err == nil {
				break
			} else if err = b.caseInGroupZone(cq); err == nil {
				break
			} else if err = b.caseInGroupRegionAttr(cq); err == nil {
				break
			} else if err = b.caseInGroupRegion(cq); err == nil {
				break
			} else {
				log.Println("dati callback incorretti")
			}
		}
	}
}

// Updates data from pcm-dpc repository
func updateData(nazione *[]covidgraphs.NationData, regioni *[]covidgraphs.RegionData, province *[]covidgraphs.ProvinceData, note *[]covidgraphs.NoteData) func() {
	return func() {
		log.Println("Updating data...")
		mutex.Lock()

		covidgraphs.DeleteAllPlots(workingDirectory + imageFolder)

		ptrNazione, err := covidgraphs.GetNation()
		if err != nil {
			log.Println("errore nell'aggiornamento dei dati nazione")
			log.Println(err)
		}
		*nazione = *ptrNazione

		ptrRegioni, err := covidgraphs.GetRegions()
		if err != nil {
			log.Println("errore nell'aggiornamento dei dati regione")
			log.Println(err)
		}
		*regioni = *ptrRegioni

		ptrProvince, err := covidgraphs.GetProvinces()
		if err != nil {
			log.Println("errore nell'aggiornamento dei dati province")
			log.Println(err)
		}
		*province = *ptrProvince

		ptrNote, err := covidgraphs.GetNotes()
		if err != nil {
			log.Println("errore nell'aggiornamento dei dati note")
			log.Println(err)
		}
		*note = *ptrNote
		mutex.Unlock()
	}
}
