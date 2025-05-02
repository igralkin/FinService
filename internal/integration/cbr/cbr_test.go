package cbr

import (
	"os"
	"testing"
)

func TestParseXMLResponse(t *testing.T) {
	// Чтение заранее сохраненного XML-ответа (мока)
	data, err := os.ReadFile("testdata/sample_response.xml")
	if err != nil {
		t.Fatalf("Ошибка чтения мок-файла: %v", err)
	}

	rate, err := parseXMLResponse(data)
	if err != nil {
		t.Errorf("Ожидался корректный парсинг, но получена ошибка: %v", err)
	}

	expected := 16.00 // поставь актуальную ставку из sample_response.xml
	if rate != expected {
		t.Errorf("Ожидалась ставка %.2f, но получено %.2f", expected, rate)
	}
}
