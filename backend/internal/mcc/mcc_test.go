package mcc

import "testing"

func TestDetectCategory(t *testing.T) {
	tests := []struct {
		name     string
		mccCode  string
		expected CategoryGroup }{
		{"Проверка супермаркета", "5411", GroupGroceries},
		{"Проверка такси", "4121", GroupTaxi},
		{"Проверка ЖКХ", "4900", GroupUtilities},
		{"Проверка кода с пробелами", "5814", GroupRestaurants},
		{"Проверка неизвестного кода", "0000", GroupUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectCategory(tt.mccCode)
			if result.Group != tt.expected {
				t.Errorf("Для кода %s ожидали группу %v, но получили %v", tt.mccCode, tt.expected, result.Group)
			}
		})
	}
}
