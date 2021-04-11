package main

import (
	"github.com/NicoNex/echotron"
	"log"
)

func (b *bot) inGroupTextNation(chatId int64) {
	buttons, err := b.buttonsConfrontoNazioneGroups(0)
	if err != nil {
		log.Println(err)
		return
	}
	b.lastGroupAttrIndex = 0
	b.SendMessageWithKeyboard("❗️❕<b>Dati nazione</b> ❕❗", chatId, buttons, echotron.PARSE_HTML)
}

func (b *bot) inGroupTextRegions(chatId int64) {
	buttons, err := b.buttonsZonesGroups(0)
	if err != nil {
		log.Println(err)
		return
	}

	b.lastZoneIndex = 0
	b.SendMessageWithKeyboard("❗️❕<b>Dati regione</b> ❕❗", chatId, buttons, echotron.PARSE_HTML)
}

func (b *bot) inGroupTextProvinces(chatId int64) {
	buttons, err := b.buttonsZonesGroupsP(0)
	if err != nil {
		log.Println(err)
		return
	}

	b.lastGroupProvinceIndex = 0
	b.lastGroupRegionIndex = 0
	b.lastZoneIndex = 0
	b.SendMessageWithKeyboard("❗️❕<b>Dati provincia</b> ❕❗", chatId, buttons, echotron.PARSE_HTML)
}
