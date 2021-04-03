package main

import (
	"encoding/json"
	"github.com/NicoNex/echotron"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

// Creates bot data folders if they don't exist
func initFolders() {
	currentPath, _ := os.Getwd()
	workingDirectory = currentPath + botDataDirectory
	os.MkdirAll(workingDirectory, 0755)
	os.MkdirAll(workingDirectory+imageFolder, 0755)
	os.MkdirAll(workingDirectory+logsFolder, 0755)
}

// Checks if a string is found in national fields selected for comparison
func (b *bot) isStringFoundInRegionChoices(str string) bool {
	for _, j := range b.choicesConfrontoRegione {
		if j == strings.ToLower(str) {
			return true
		}
	}
	return false
}

// Checks if a string is found in national fields selected for comparison
func (b *bot) isStringFoundInNationChoices(str string) bool {
	for _, j := range b.choicesConfrontoNazione {
		if j == strings.ToLower(str) {
			return true
		}
	}
	return false
}

// Logs operations to a file named with telegram username/name_surname
func writeOperation(cq *echotron.CallbackQuery) {
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

// Sorts regional fields selected for comparison
func (b *bot) getSortedChoicesConfrontoRegione() []string {
	tempChoices := make([]string, len(b.choicesConfrontoRegione))
	copy(tempChoices, b.choicesConfrontoRegione)
	sort.Strings(tempChoices)

	return tempChoices
}

// Sorts national fields selected for comparison
func (b *bot) getSortedChoicesConfrontoNazione() []string {
	tempChoices := make([]string, len(b.choicesConfrontoNazione))
	copy(tempChoices, b.choicesConfrontoNazione)
	sort.Strings(tempChoices)

	return tempChoices
}
