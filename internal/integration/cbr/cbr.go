package cbr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beevik/etree"
	log "github.com/sirupsen/logrus"
)

func buildSOAPRequest() string {
	fromDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	toDate := time.Now().Format("2006-01-02")

	log.WithFields(log.Fields{
		"from": fromDate,
		"to":   toDate,
	}).Info("Building SOAP request to the Central Bank of Russia (30-day period)")

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
        <soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
            <soap12:Body>
                <KeyRate xmlns="http://web.cbr.ru/">
                    <fromDate>%s</fromDate>
                    <ToDate>%s</ToDate>
                </KeyRate>
            </soap12:Body>
        </soap12:Envelope>`, fromDate, toDate)
}

func sendRequest(soapRequest string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("POST", "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx", bytes.NewBuffer([]byte(soapRequest)))
	if err != nil {
		log.WithError(err).Error("Failed to create HTTP request")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

	log.Info("Sending SOAP request to the Central Bank of Russia...")
	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("Error while sending request to the Central Bank of Russia")
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("Failed to read response from the Central Bank of Russia")
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	log.Info("Response from the Central Bank of Russia received successfully")
	return rawBody, nil
}

func parseXMLResponse(rawBody []byte) (float64, error) {
	log.Debug("Starting XML response parsing...")

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(rawBody); err != nil {
		log.WithError(err).Error("Failed to parse XML")
		return 0, fmt.Errorf("ошибка парсинга XML: %v", err)
	}

	krElements := doc.FindElements("//diffgram/KeyRate/KR")
	if len(krElements) == 0 {
		log.Warn("KeyRate elements not found in XML response")
		return 0, errors.New("данные по ставке не найдены")
	}

	rateElement := krElements[0].FindElement("./Rate")
	if rateElement == nil {
		log.Warn("Rate element is missing")
		return 0, errors.New("тег Rate отсутствует")
	}

	var rate float64
	if _, err := fmt.Sscanf(rateElement.Text(), "%f", &rate); err != nil {
		log.WithError(err).Error("Failed to convert rate value")
		return 0, fmt.Errorf("ошибка конвертации: %v", err)
	}

	log.WithField("rate", rate).Info("Key rate successfully extracted")
	return rate, nil
}

// GetCentralBankRate получает ключевую ставку ЦБ РФ и прибавляет маржу банка
func GetCentralBankRate() (float64, error) {
	request := buildSOAPRequest()
	response, err := sendRequest(request)
	if err != nil {
		return 0, err
	}

	rate, err := parseXMLResponse(response)
	if err != nil {
		return 0, err
	}

	finalRate := rate + 5
	log.WithFields(log.Fields{
		"baseRate":   rate,
		"bankMargin": 5,
		"finalRate":  finalRate,
	}).Info("Final rate with bank margin calculated")

	return finalRate, nil
}
