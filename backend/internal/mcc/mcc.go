package mcc

import "strings"

type Category struct {
	Name  string
	Group string
}

var mccRegistry = map[string]Category{
	"5411": {Name: "Бакалейные магазины, супермаркеты", Group: "Супермаркеты"},
	"5499": {Name: "Различные продовольственные магазины", Group: "Супермаркеты"},
	"5812": {Name: "Места общественного питания, рестораны", Group: "Рестораны"},
	"5814": {Name: "Рестораны быстрого обслуживания (Фастфуд)", Group: "Фастфуд"},
	"5813": {Name: "Бары, коктейль-холлы, дискотеки", Group: "Бары"},
	"4111": {Name: "Пригородные поезда, общественный транспорт", Group: "Транспорт"},
	"4121": {Name: "Такси и каршеринг", Group: "Такси"},
	"5912": {Name: "Аптеки", Group: "Аптеки"},
	"5211": {Name: "Материалы для строительства, ремонта", Group: "Дом и ремонт"},
	"5541": {Name: "Заправочные станции (АЗС)", Group: "Топливо"},
}

func DetectCategory(code string) Category {
	cleanCode := strings.TrimSpace(code)

	category, exists := mccRegistry[cleanCode]

	if !exists {
		return Category{
			Name:  "Неизвестный МСС-код",
			Group: "Другое",
		}
	}
	return category
}
