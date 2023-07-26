package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strings"
)

type Viability struct {
	NumeroTelefone string `json:"numero_telefone"`
	Cep string `json:"cep"`
	Numero_casa string `json:"numero_casa"`
	Tipo_viabilidade string `json:"tipo_viabilidade"`
	Status_viabilidade bool `json:"status_viabilidade"`
	Viability_oi string `json:"viability_oi"`
}


// Function to check viability for an address in MongoDB
func checkViabilityForAddress(cep string, number string, phoneNumber string) (*Viability, error) {
	fmt.Println("Checking viability for address:", cep, number)

	apiURL := fmt.Sprintf("https://test.com.br/viability/v3/check?zipcode=%s&number=%s", cep, number)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Basic Example token")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse map[string]interface{}
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	responseAPI := apiResponse["oi"].(map[string]interface{})


	viabilidadeOi := responseAPI["status_hp"].(string)

	status := responseAPI["status_viability"].(bool)
	tipo_viabilidade := responseAPI["type_viability"].(string)
	cep, ok := responseAPI["cep"].(string)
	if !ok {
		return nil, nil
	}

	return &Viability{
		NumeroTelefone:    phoneNumber,
		Cep:               cep,
		Numero_casa:       number,
		Tipo_viabilidade:  tipo_viabilidade,
		Status_viabilidade: status,
		Viability_oi:      viabilidadeOi,
	}, nil
}

func main() {
	inputCSVFilePath := "./input.csv"
	outputCSVFilePath := "./file2.csv"

	inputFile, err := os.Open(inputCSVFilePath)
	if err != nil {
		log.Fatalf("Error opening input CSV file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputCSVFilePath)
	if err != nil {
		log.Fatalf("Error creating output CSV file: %v", err)
	}
	defer outputFile.Close()

	csvReader := csv.NewReader(inputFile)
	csvWriter := csv.NewWriter(outputFile)
	_, err = csvReader.Read()
	if err != nil {
		log.Fatalf("Error reading CSV header: %v", err)
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error reading CSV record: %v", err)
		}

		cep := strings.TrimSpace(record[42])
		number := strings.TrimSpace(record[43])
		phoneNumber := strings.TrimSpace(record[10])

		viabilityOi, err := checkViabilityForAddress(cep, number, phoneNumber)
		if err != nil || viabilityOi == nil{
			fmt.Printf("Error checking viability for address %s, %s: %v", cep, number, err)
		} else {
			fmt.Println("Viability for address:", viabilityOi)
	
			csvWriter.Write([]string{phoneNumber, cep, number, viabilityOi.Tipo_viabilidade, viabilityOi.Viability_oi})
			csvWriter.Flush()
		}
	}

	fmt.Printf("Results have been written to: %s\n", outputCSVFilePath)
}
