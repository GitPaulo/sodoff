package cmd

import (
	"fmt"
	"os"
	"sodoff/api"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	nr "github.com/martinsirbe/go-national-rail-client/nationalrail"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

var (
	continuous         bool
	interval           int
	departureStation   string
	destinationStation string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "sodoff",
	Short: "Sodoff is a CLI tool for checking train cancellations",
	Long: `Sodoff is a CLI tool for checking train cancellations and alerting the user for any future
cancellations.`,
	Run: runRootCmd,
}

func init() {
	rootCmd.Flags().BoolVarP(&continuous, "continuous", "c", false, "Continuously check for updates")
	rootCmd.Flags().IntVarP(&interval, "interval", "i", 5, "Polling interval in seconds")
	rootCmd.Flags().StringVarP(&departureStation, "from", "f", "", "Departure station CRS code or name")
	rootCmd.Flags().StringVarP(&destinationStation, "to", "t", "", "Destination station CRS code or name")
}

func checkAccessToken() bool {
	const tokenEnvVar = "NR_ACCESS_TOKEN"
	const tokenURL = "https://www.nationalrail.co.uk/developers/"
	token := os.Getenv(tokenEnvVar)
	if token == "" {
		fmt.Println(color.RedString("ERROR: National Rail API access token not found!"))
		fmt.Println("Please set the environment variable", color.CyanString(tokenEnvVar), "with your National Rail API access token.")
		fmt.Println("You can obtain a token from the following link:")
		fmt.Println(color.CyanString(tokenURL))

		// Open the URL in the default browser
		err := open.Run(tokenURL)
		if err != nil {
			fmt.Println("Please visit the URL to obtain your token:", tokenURL)
		}
		return false
	}
	return true
}

func runRootCmd(cmd *cobra.Command, args []string) {
	if !checkAccessToken() {
		return
	}

	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")

	departureStation := validateStationInput(from, "Select Departure Station")
	if departureStation == "" {
		fmt.Printf("Invalid departure station: %s\n", from)
		return
	}

	destinationStation := validateStationInput(to, "Select Destination Station")
	if destinationStation == "" {
		fmt.Printf("Invalid destination station: %s\n", to)
		return
	}

	if continuous {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		for {
			select {
			case <-ticker.C:
				display(departureStation, destinationStation)
			}
		}
	} else {
		display(departureStation, destinationStation)
	}
}

func validateStationInput(station, promptLabel string) string {
	if station == "" {
		return selectStation(promptLabel)
	}

	validStation := getStationCode(station)
	if validStation == "" {
		fmt.Printf("Invalid station: %s\n", station)
		return selectStation(promptLabel)
	}

	return validStation
}

func getStationCode(station string) string {
	station = strings.ToUpper(station)
	if _, exists := nr.StationCodeToNameMap[nr.CRSCode(station)]; exists {
		return station
	}

	stations, err := api.SearchStations(station)
	if err != nil {
		fmt.Printf("Failed to search stations: %v\n", err)
		return ""
	}

	if len(stations) > 0 {
		return stations[0].CRS
	}
	return ""
}

func selectStation(promptLabel string) string {
	for {
		prompt := promptui.Prompt{
			Label: promptLabel,
			Validate: func(input string) error {
				if len(input) < 2 {
					return fmt.Errorf("search query must be at least 2 characters")
				}
				return nil
			},
		}

		searchQuery, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return ""
		}

		stations, err := api.SearchStations(searchQuery)
		if err != nil {
			fmt.Printf("Failed to search stations: %v\n", err)
			return ""
		}

		if len(stations) == 0 {
			fmt.Println("No stations found, please try again.")
			continue
		}

		stationNames := make([]string, len(stations))
		stationMap := make(map[string]nr.CRSCode)
		for i, station := range stations {
			stationNames[i] = station.Name
			stationMap[station.Name] = nr.CRSCode(station.CRS)
		}

		selectPrompt := promptui.Select{
			Label:             "Select Station",
			Items:             stationNames,
			StartInSearchMode: true,
			Searcher: func(input string, index int) bool {
				return fuzzySearch(input, stationNames[index])
			},
		}

		_, stationName, err := selectPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return ""
		}

		return string(stationMap[stationName])
	}
}

func fuzzySearch(input, item string) bool {
	input = strings.ToLower(input)
	item = strings.ToLower(item)
	for len(input) <= len(item) {
		if strings.HasPrefix(item, input) {
			return true
		}
		item = item[1:]
	}
	return false
}

func display(departureStation, destinationStation string) {
	departureBoard, err := api.GetDeparturesBoard(nr.CRSCode(departureStation))
	if err != nil {
		fmt.Printf("Error fetching station board for %s: %v\n", departureStation, err)
		return
	}
	fmt.Println(displayDepartureBoard(departureStation, departureBoard, "Departure Board"))

	arrivalBoard, err := api.GetArrivalsBoard(nr.CRSCode(destinationStation))
	if err != nil {
		fmt.Printf("Error fetching station board for %s: %v\n", destinationStation, err)
		return
	}
	fmt.Println(displayArrivalBoard(destinationStation, arrivalBoard, "Arrivals Board"))
}

func displayDepartureBoard(station string, board *nr.StationBoard, boardTitle string) string {
	stationName := ""
	if name, exists := nr.StationCodeToNameMap[nr.CRSCode(station)]; exists {
		stationName = name
	}

	titleFigure := figure.NewColorFigure(
		fmt.Sprintf("%s - %s [%s]", boardTitle, stationName, station),
		"short",
		"green",
		true,
	)
	titleFigure.Print()

	var builder strings.Builder
	builder.WriteString("=========================================================================================================\n")
	builder.WriteString(fmt.Sprintf("%-10s %-30s %-10s %-10s %-20s %-40s\n", "STD", "Destination", "Platform", "Status", "ETD", "Operator"))
	builder.WriteString("---------------------------------------------------------------------------------------------------------\n")

	for _, service := range board.TrainServices {
		std := service.STD

		destination := ""
		if service.Destination != nil {
			destination = service.Destination.Name
		}

		platform := ""
		if service.Platform != nil {
			platform = *service.Platform
		}

		status := getStatus(service)
		statusColor := getColor(status)

		etd := service.ETD
		if etd == "" {
			etd = "N/A"
		}

		row := fmt.Sprintf("%-10s %-30s %-10s ", std, destination, platform)
		builder.WriteString(row)
		statusColor.Fprintf(&builder, "%-10s %-20s %-40s\n", status, etd, service.Operator)
	}

	builder.WriteString("=========================================================================================================\n")

	return builder.String()
}

func displayArrivalBoard(station string, board *nr.StationBoard, boardTitle string) string {
	stationName := ""
	if name, exists := nr.StationCodeToNameMap[nr.CRSCode(station)]; exists {
		stationName = name
	}

	titleFigure := figure.NewColorFigure(
		fmt.Sprintf("%s - %s [%s]", boardTitle, stationName, station),
		"short",
		"green",
		true,
	)
	titleFigure.Print()

	var builder strings.Builder
	builder.WriteString("=========================================================================================================\n")
	builder.WriteString(fmt.Sprintf("%-10s %-30s %-10s %-10s %-20s %-40s\n", "STA", "Origin", "Platform", "Status", "ETA", "Operator"))
	builder.WriteString("---------------------------------------------------------------------------------------------------------\n")

	for _, service := range board.TrainServices {
		sta := "N/A"
		if service.STA != nil {
			sta = *service.STA
		}

		origin := ""
		if service.Origin != nil {
			origin = service.Origin.Name
		}

		platform := ""
		if service.Platform != nil {
			platform = *service.Platform
		}

		status := getStatus(service)
		statusColor := getColor(status)

		eta := "N/A"
		if service.ETA != nil {
			eta = *service.ETA
		}

		row := fmt.Sprintf("%-10s %-30s %-10s ", sta, origin, platform)
		builder.WriteString(row)
		statusColor.Fprintf(&builder, "%-10s %-20s %-40s\n", status, eta, service.Operator)
	}

	builder.WriteString("=========================================================================================================\n")

	return builder.String()
}

func getStatus(service *nr.TrainService) string {
	eta := "N/A"
	if service.ETA != nil {
		eta = *service.ETA
	}
	etd := service.ETD
	if etd == "" {
		etd = "N/A"
	}
	if service.STD == "Cancelled" || service.ETD == "Cancelled" || eta == "Cancelled" || etd == "Cancelled" {
		return "Cancelled"
	}
	if service.DelayReason != nil {
		return "Delayed"
	}
	return "On time"
}

func getColor(status string) *color.Color {
	switch status {
	case "Cancelled":
		return color.New(color.FgRed)
	case "Delayed":
		return color.New(color.FgYellow)
	default:
		return color.New(color.FgGreen)
	}
}
