package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Application struct {
	config *Config
}

type Config struct {
}

// Shift - barebone of Schedule, contains informations who is working - shift span
type Shift struct {
	EmployeeID string    // ID of the employee working this shift
	Start      time.Time // Start time of the shift in HH:MM format
	End        time.Time // End time of the shift in HH:MM format
}

type Schedule struct {
	Shifts []Shift
}

type Employee struct {
	ID   int // Unique identifier for the employee
	Name string
}

func main() {

	schedule, err := LoadScheduleFromCSV("data.csv")
	if err != nil {
		fmt.Printf("Błąd wczytywania harmonogramu: %v\n", err)
		return
	}
	fmt.Println("Harmonogram pracowników:")
	for _, shift := range schedule.Shifts {
		fmt.Printf("Pracownik: %s, Start: %s, Koniec: %s\n",
			shift.EmployeeID, shift.Start.Format(time.RFC822), shift.End.Format(time.RFC822))
	}
}

func LoadScheduleFromCSV(filePath string) (*Schedule, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // pozwala na różne długości wierszy

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("błąd nagłówka CSV: %w", err)
	}

	var schedule Schedule

	// Mapujemy indeksy kolumn do EmployeeID
	employeeIDs := []string{}
	for i := 1; i < len(header); i += 2 {
		col := strings.TrimSuffix(header[i], "_Start")
		employeeIDs = append(employeeIDs, col)
	}

	layoutDate := "2006-02-01"
	layoutTime := "15:04"

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("błąd rekordu CSV: %w", err)
		}

		dateStr := record[0]
		baseDate, err := time.Parse(layoutDate, dateStr)
		if err != nil {
			return nil, fmt.Errorf("błędna data: %s", dateStr)
		}

		for i, empID := range employeeIDs {
			startStr := strings.TrimSpace(record[i*2+1])
			endStr := strings.TrimSpace(record[i*2+2])

			if startStr == "0:00" || endStr == "0:00" {
				continue // brak zmiany
			}

			startClock, err := time.Parse(layoutTime, startStr)
			if err != nil {
				return nil, fmt.Errorf("błędny start: %s", startStr)
			}

			endClock, err := time.Parse(layoutTime, endStr)
			if err != nil {
				return nil, fmt.Errorf("błędny koniec: %s", endStr)
			}

			// Złóż pełne daty z godzinami
			start := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
				startClock.Hour(), startClock.Minute(), 0, 0, time.Local)

			end := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
				endClock.Hour(), endClock.Minute(), 0, 0, time.Local)

			// Jeśli koniec jest wcześniej niż start – nocna zmiana
			if end.Before(start) || end.Equal(start) {
				end = end.Add(24 * time.Hour)
			}

			shift := Shift{
				EmployeeID: empID,
				Start:      start,
				End:        end,
			}

			schedule.Shifts = append(schedule.Shifts, shift)
		}
	}

	return &schedule, nil
}
