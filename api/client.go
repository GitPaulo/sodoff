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

func GetArrivalsBoard(crs nr.CRSCode) (*nr.StationBoard, error) {
	// Requires: env var NR_ACCESS_TOKEN
	client, err := nr.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	board, err := client.GetArrivals(crs, nr.NumRowsOpt(10))
	if err != nil {
		return nil, fmt.Errorf("failed to get arrivals board: %v", err)
	}

	// No services found. Alert the user.
	if len(board.TrainServices) == 0 {
		fmt.Println("No train services found for station:", nr.StationCodeToNameMap[crs])
	}

	return board, nil
}

func GetDeparturesBoard(crs nr.CRSCode) (*nr.StationBoard, error) {
	// Requires: env var NR_ACCESS_TOKEN
	client, err := nr.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	board, err := client.GetDepartures(crs, nr.NumRowsOpt(10))
	if err != nil {
		return nil, fmt.Errorf("failed to get departures board: %v", err)
	}

	// No services found. Alert the user.
	if len(board.TrainServices) == 0 {
		fmt.Println("No train services found for station:", nr.StationCodeToNameMap[crs])
	}

	return board, nil
}
