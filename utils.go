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
	workingDirectory = currentPath + "/" + botDataDirectory
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
func writeOperation(update *echotron.Update, folder string) {
	data, err := json.Marshal(update)
	if err != nil {
		log.Println("Error marshaling logs: ", err)
		return
	}

	var filename string

	if update.CallbackQuery != nil {
		if update.CallbackQuery.Message.Chat.Type == "private" {
			if update.CallbackQuery.Message.Chat.Username == "" {
				filename = folder + update.CallbackQuery.Message.Chat.FirstName + "_" + update.CallbackQuery.Message.Chat.LastName + ".txt"
			} else {
				filename = folder + update.CallbackQuery.Message.Chat.Username + ".txt"
			}
		} else {
			filename = folder + update.CallbackQuery.Message.Chat.Title + ".txt"
		}

	} else if update.Message != nil {
		if update.Message.Chat.Type == "private" {
			if update.Message.Chat.Username == "" {
				filename = folder + update.Message.Chat.FirstName + "_" + update.Message.Chat.LastName + ".txt"
			} else {
				filename = folder + update.Message.Chat.Username + ".txt"
			}
		} else {
			filename = folder + update.Message.Chat.Title + ".txt"
		}

	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		return
	}

	dataString := time.Now().Format("2006-01-02T15:04:05") + string(data[:])
	_, err = f.WriteString(dataString + "\n")
	if err != nil {
		log.Println(err)
		return
	}
	err = f.Close()
	if err != nil {
		log.Println(err)
		return
	}
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
