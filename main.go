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

func TimeDiffBetweenShifts(start, end *Shift) (time.Duration, error) {
	if start == nil || end == nil {
		return 0, fmt.Errorf("nie można obliczyć różnicy czasu między pustymi zmianami")
	}
	if start.End.Before(start.Start) || end.End.Before(end.Start) {
		return 0, fmt.Errorf("nieprawidłowe czasy rozpoczęcia lub zakończenia zmian")
	}
	return end.End.Sub(start.Start), nil
}

type Schedule struct {
	Shifts []Shift
}

func (s *Schedule) WorkTimeInRange(fromDate, toDate time.Time, workerID string) (time.Duration, error) {
	if fromDate.After(toDate) {
		return 0, fmt.Errorf("data początkowa nie może być późniejsza niż data końcowa")
	}

	var totalDuration time.Duration
	for _, shift := range s.Shifts {
		if shift.EmployeeID == workerID && shift.Start.Before(toDate) && shift.End.After(fromDate) {
			start := max(shift.Start, fromDate)
			end := min(shift.End, toDate)
			totalDuration += end.Sub(start)
		}
	}
	return totalDuration, nil
}
func max(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
func min(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

type Employee struct {
	ID   int // Unique identifier for the employee
	Name string
}

func main() {

	// Check work time in range

	schedule := &Schedule{
		Shifts: []Shift{
			{EmployeeID: "123", Start: time.Date(2023, 10, 1, 8, 0, 0, 0, time.Local), End: time.Date(2023, 10, 1, 16, 0, 0, 0, time.Local)},
			{EmployeeID: "123", Start: time.Date(2023, 10, 2, 8, 0, 0, 0, time.Local), End: time.Date(2023, 10, 2, 16, 0, 0, 0, time.Local)},
			{EmployeeID: "456", Start: time.Date(2023, 10, 1, 16, 0, 0, 0, time.Local), End: time.Date(2023, 10, 1, 24, 0, 0, 0, time.Local)},
			{EmployeeID: "456", Start: time.Date(2023, 10, 2, 16, 0, 0, 0, time.Local), End: time.Date(2023, 10, 2, 24, 0, 0, 0, time.Local)},
		},
	}
	fromDate := time.Date(2023, 10, 1, 0, 0, 0, 0, time.Local)
	toDate := time.Date(2023, 10, 2, 0, 0, 0, 0, time.Local)
	workTime, err := schedule.WorkTimeInRange(fromDate, toDate, "123")
	if err != nil {
		fmt.Printf("Błąd obliczania czasu pracy: %v\n", err)
		return
	}
	fmt.Printf("Czas pracy pracownika 123 w podanym zakresie: %v\n", workTime)

	// Check time difference between shifts
	// shift1 := &Shift{
	// 	EmployeeID: "123",
	// 	Start:      time.Date(2023, 10, 1, 8, 0, 0, 0, time.Local),
	// 	End:        time.Date(2023, 10, 1, 16, 0, 0, 0, time.Local),
	// }
	// shift2 := &Shift{
	// 	EmployeeID: "456",
	// 	Start:      time.Date(2023, 10, 1, 16, 0, 0, 0, time.Local),
	// 	End:        time.Date(2023, 10, 1, 24, 0, 0, 0, time.Local), // Nocna zmiana
	// }
	// diff, err := TimeDiffBetweenShifts(shift1, shift2)
	// if err != nil {
	// 	fmt.Printf("Błąd obliczania różnicy czasu: %v\n", err)
	// 	return
	// }
	// fmt.Printf("Różnica czasu między zmianami: %v\n", diff)

	// Check csv file loading

	// schedule, err := LoadScheduleFromCSV("data.csv")
	// if err != nil {
	// 	fmt.Printf("Błąd wczytywania harmonogramu: %v\n", err)
	// 	return
	// }
	// fmt.Println("Harmonogram pracowników:")
	// for _, shift := range schedule.Shifts {
	// 	fmt.Printf("Pracownik: %s, Start: %s, Koniec: %s\n",
	// 		shift.EmployeeID, shift.Start.Format(time.RFC822), shift.End.Format(time.RFC822))
	// }
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
