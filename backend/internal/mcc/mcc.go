package mcc

import "strings"

type CategoryGroup string

const (
	GroupGroceries     CategoryGroup = "Продукты"
	GroupRestaurants   CategoryGroup = "Кафе и рестораны"
	GroupTaxi          CategoryGroup = "Такси"
	GroupTransport     CategoryGroup = "Транспорт"
	GroupAuto          CategoryGroup = "Автомобиль"
	GroupClothes       CategoryGroup = "Одежда и обувь"
	GroupHealth        CategoryGroup = "Здоровье и аптеки"
	GroupHome          CategoryGroup = "Дом и ремонт"
	GroupUtilities     CategoryGroup = "Коммунальные платежи"
	GroupElectronics   CategoryGroup = "Техника и гаджеты"
	GroupEntertainment CategoryGroup = "Развлечения и досуг"
	GroupServices      CategoryGroup = "Услуги и связь"
	GroupTransfers     CategoryGroup = "Переводы и наличные"
	GroupUnknown       CategoryGroup = "Разное"
)

type Category struct {
	Name  string
	Group CategoryGroup
}

var mccRegistry = map[string]Category{
	"5411": {Name: "Супермаркеты и бакалея", Group: GroupGroceries},
	"5499": {Name: "Продовольственные магазины", Group: GroupGroceries},
	"5812": {Name: "Рестораны и кафе", Group: GroupRestaurants},
	"5814": {Name: "Фастфуд", Group: GroupRestaurants},
	"5813": {Name: "Бары и клубы", Group: GroupRestaurants},
	"4121": {Name: "Такси и каршеринг", Group: GroupTaxi},
	"4111": {Name: "Общественный транспорт и метро", Group: GroupTransport},
	"5541": {Name: "Заправочные станции (АЗС)", Group: GroupAuto},
	"5651": {Name: "Магазины одежды", Group: GroupClothes},
	"5661": {Name: "Магазины обуви", Group: GroupClothes},
	"5732": {Name: "Электроника и бытовая техника", Group: GroupElectronics},
	"5300": {Name: "Маркетплейсы (Ozon, WB, Яндекс.Маркет)", Group: GroupGroceries},
	"4900": {Name: "Коммунальные услуги и ЖКХ", Group: GroupUtilities},
	"5211": {Name: "Строительные материалы и ремонт", Group: GroupHome},
	"4814": {Name: "Мобильная связь и интернет", Group: GroupServices},
	"5912": {Name: "Аптеки", Group: GroupHealth},
	"8011": {Name: "Медицинские услуги и клиники", Group: GroupHealth},
	"7832": {Name: "Кинотеатры и билеты", Group: GroupEntertainment},
	"7997": {Name: "Спортзалы и фитнес", Group: GroupEntertainment},
	"4829": {Name: "Денежные переводы", Group: GroupTransfers},
	"6011": {Name: "Снятие наличных", Group: GroupTransfers},
}

func DetectCategory(code string) Category {
	cleanCode := strings.TrimSpace(code)

	if category, exists := mccRegistry[cleanCode]; exists {
		return category
	}

	return Category{
		Name:  "Неизвестный MCC-код",
		Group: GroupUnknown,
	}
}
