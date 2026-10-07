package mcc

import "testing"

func TestCategory(t *testing.T) {
	result := DetectCategory("5411")
	if result.Group != "Супермаркеты" {
		t.Errorf("Ожидали Супермаркеты, получили %s", result.Group)
	}
}
