package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "regexp"
    "io"
)

type ViaCEP struct {
    Localidade string `json:"localidade"`
    Erro       bool   `json:"erro,omitempty"`
}

type WeatherAPI struct {
    Current struct {
        TempC float64 `json:"temp_c"`
    } `json:"current"`
}

type Temperature struct {
    TempC float64 `json:"temp_C"`
    TempF float64 `json:"temp_F"`
    TempK float64 `json:"temp_K"`
}

func isValidCEP(cep string) bool {
    match, _ := regexp.MatchString(`^\d{8}$`, cep)
    return match
}

func getCityFromCEP(cep string) (string, int, error) {
    resp, err := http.Get(fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep))
    if err != nil {
        return "", http.StatusInternalServerError, err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    var data ViaCEP
    err = json.Unmarshal(body, &data)
    if err != nil {
        return "", http.StatusInternalServerError, err
    }
    if data.Erro {
        return "", http.StatusNotFound, fmt.Errorf("can not find zipcode")
    }
    return data.Localidade, http.StatusOK, nil
}

func getTemperature(city string) (float64, error) {
    apiKey := os.Getenv("f258c491f41b4a1f8c405103252507")
    url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", apiKey, city)

    res, err := http.Get(url)
    if err != nil {
        return 0, err
    }
    defer res.Body.Close()

    body, _ := io.ReadAll(res.Body)

    var data WeatherAPI
    err = json.Unmarshal(body, &data)
    if err != nil {
        return 0, err
    }
    return data.Current.TempC, nil
}

func temperatureHandler(w http.ResponseWriter, r *http.Request) {
    cep := r.URL.Query().Get("cep")

    if !isValidCEP(cep) {
        http.Error(w, `{"message": "invalid zipcode"}`, http.StatusUnprocessableEntity)
        return
    }

    city, code, err := getCityFromCEP(cep)
    if err != nil {
        http.Error(w, fmt.Sprintf(`{"message": "%s"}`, err.Error()), code)
        return
    }

    tempC, err := getTemperature(city)
    if err != nil {
        http.Error(w, `{"message": "failed to get temperature"}`, http.StatusInternalServerError)
        return
    }

    result := Temperature{
        TempC: tempC,
        TempF: tempC * 1.8 + 32,
        TempK: tempC + 273,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

func main() {
    http.HandleFunc("/weather", temperatureHandler)
    
    log.Printf("Servidor respondendo na porta 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}