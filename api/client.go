package api

import (
	"fmt"
	"strings"

	nr "github.com/martinsirbe/go-national-rail-client/nationalrail"
)

func SearchStations(query string) ([]nr.Location, error) {
	query = strings.ToLower(query)
	var results []nr.Location
	for code, name := range nr.StationCodeToNameMap {
		if strings.Contains(strings.ToLower(name), query) {
			results = append(results, nr.Location{
				Name: name,
				CRS:  string(code),
				Type: "train_station",
			})
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no stations found for query: %s", query)
	}

	return results, nil
}

// GetArrivalsBoard fetches the arrivals board for a given CRS code with custom options.
func GetArrivalsBoard(crs nr.CRSCode, numRows int, timeWindowMinutes int) (*nr.StationBoard, error) {
	client, err := nr.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	board, err := client.GetArrivalsWithDetails(crs, nr.NumRowsOpt(numRows), nr.TimeWindowMinutesOpt(timeWindowMinutes))
	if err != nil {
		return nil, fmt.Errorf("failed to get arrival board: %v", err)
	}

	if len(board.TrainServices) == 0 {
		fmt.Println("No train services found for station:", nr.StationCodeToNameMap[crs])
	}

	return board, nil
}

// GetDeparturesBoard fetches the departures board for a given CRS code with custom options.
func GetDeparturesBoard(crs nr.CRSCode, numRows int, timeWindowMinutes int) (*nr.StationBoard, error) {
	client, err := nr.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	board, err := client.GetDeparturesWithDetails(crs, nr.NumRowsOpt(numRows), nr.TimeWindowMinutesOpt(timeWindowMinutes))
	if err != nil {
		return nil, fmt.Errorf("failed to get departure board: %v", err)
	}

	if len(board.TrainServices) == 0 {
		fmt.Println("No train services found for station:", nr.StationCodeToNameMap[crs])
	}

	return board, nil
}
