package cmd

import (
	"fmt"
	"log"
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

const (
	defaultInterval   = 5
	defaultNumRows    = 10
	defaultTimeWindow = 60
	tokenEnvVar       = "NR_ACCESS_TOKEN"
	tokenURL          = "https://www.nationalrail.co.uk/developers/"
)

var (
	continuous         bool
	interval           int
	departureStation   string
	destinationStation string
	numRows            int
	timeWindow         int
	showJourneys       bool
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
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
	rootCmd.Flags().IntVarP(&interval, "interval", "i", defaultInterval, "Polling interval in seconds")
	rootCmd.Flags().StringVarP(&departureStation, "from", "f", "", "Departure station CRS code or name")
	rootCmd.Flags().StringVarP(&destinationStation, "to", "t", "", "Destination station CRS code or name")
	rootCmd.Flags().IntVarP(&numRows, "rows", "r", defaultNumRows, "Number of rows to fetch (Don't change this unless you know what you're doing)")
	rootCmd.Flags().IntVarP(&timeWindow, "time-window", "w", defaultTimeWindow, "Time window in minutes")
	rootCmd.Flags().BoolVarP(&showJourneys, "show-journeys", "j", false, "Show full journey of highlighted trains")
}

func checkAccessToken() bool {
	token := os.Getenv(tokenEnvVar)
	if token == "" {
		fmt.Println(color.RedString("ERROR: National Rail API access token not found!"))
		fmt.Println("Please set the environment variable", color.CyanString(tokenEnvVar), "with your National Rail API access token.")
		fmt.Println("You can obtain a token from the following link:")
		fmt.Println(color.CyanString(tokenURL))

		// Open the URL in the default browser
		if err := open.Run(tokenURL); err != nil {
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
	numRows, _ := cmd.Flags().GetInt("rows")
	timeWindow, _ := cmd.Flags().GetInt("time-window")

	departureStationCRS := validateStationInput(from, "Select Departure Station")
	if departureStationCRS == "" {
		log.Printf("Invalid departure station: %s\n", from)
		return
	}

	destinationStationCRS := validateStationInput(to, "Select Destination Station")
	if destinationStationCRS == "" {
		log.Printf("Invalid destination station: %s\n", to)
		return
	}

	if continuous {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		for {
			select {
			case <-ticker.C:
				display(departureStationCRS, destinationStationCRS, numRows, timeWindow)
			}
		}
	} else {
		display(departureStationCRS, destinationStationCRS, numRows, timeWindow)
	}
}

func validateStationInput(station, promptLabel string) string {
	if station == "" {
		return selectStation(promptLabel)
	}

	validStation := getStationCode(station)
	if validStation == "" {
		log.Printf("Invalid station: %s\n", station)
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
		log.Printf("Failed to search stations: %v\n", err)
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
			log.Printf("Prompt failed %v\n", err)
			return ""
		}

		stations, err := api.SearchStations(searchQuery)
		if err != nil {
			log.Printf("Failed to search stations: %v\n", err)
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
			log.Printf("Prompt failed %v\n", err)
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

func display(departureStationCRS, destinationStationCRS string, numRows int, timeWindowMinutes int) {
	departureBoard, err := api.GetDeparturesBoard(nr.CRSCode(departureStationCRS), numRows, timeWindowMinutes)
	if err != nil {
		log.Printf("Error fetching station board for %s: %v\n", departureStationCRS, err)
		return
	}
	fmt.Println(displayDepartureBoard(departureStationCRS, destinationStationCRS, departureBoard, "Departure Board"))

	arrivalBoard, err := api.GetArrivalsBoard(nr.CRSCode(destinationStationCRS), numRows, timeWindowMinutes)
	if err != nil {
		log.Printf("Error fetching station board for %s: %v\n", destinationStationCRS, err)
		return
	}
	fmt.Println(displayArrivalBoard(destinationStationCRS, departureStationCRS, arrivalBoard, "Arrivals Board"))
}

func displayDepartureBoard(departureStationCRS, destinationStationCRS string, board *nr.StationBoard, boardTitle string) string {
	departureStationName := getStationName(departureStationCRS)

	titleFigure := figure.NewColorFigure(
		fmt.Sprintf("%s - %s [%s]", boardTitle, departureStationName, departureStationCRS),
		"short",
		"green",
		true,
	)
	titleFigure.Print()

	var builder strings.Builder
	var reasonsBuilder strings.Builder
	var journeys []string

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

		highlight := containsIntermediateStation(service, departureStationCRS, destinationStationCRS)
		if highlight {
			color.New(color.BgBlue).Fprint(&builder, row)
			statusColor.Fprintf(&builder, "%-10s %-20s %-40s\n", status, etd, service.Operator)
			journeys = append(journeys, formatJourney(service))
		} else {
			builder.WriteString(row)
			statusColor.Fprintf(&builder, "%-10s %-20s %-40s\n", status, etd, service.Operator)
		}

		if service.DelayReason != nil {
			reasonsBuilder.WriteString(fmt.Sprintf("\t‣ %s to %s - %s\n", departureStationName, destination, *service.DelayReason))
		}
	}

	builder.WriteString("=========================================================================================================\n")
	if reasonsBuilder.Len() > 0 {
		builder.WriteString("Reasons for delays/cancellations:\n")
		builder.WriteString(reasonsBuilder.String())
		builder.WriteString("=========================================================================================================\n")
	}

	if showJourneys && len(journeys) > 0 {
		builder.WriteString("Highlighted Journeys:\n")
		for _, journey := range journeys {
			builder.WriteString(journey)
		}
		builder.WriteString("=========================================================================================================\n")
	}

	return builder.String()
}

func displayArrivalBoard(arrivalStationCRS, departureStationCRS string, board *nr.StationBoard, boardTitle string) string {
	arrivalStationName := getStationName(arrivalStationCRS)

	titleFigure := figure.NewColorFigure(
		fmt.Sprintf("%s - %s [%s]", boardTitle, arrivalStationName, arrivalStationCRS),
		"short",
		"green",
		true,
	)
	titleFigure.Print()

	var builder strings.Builder
	var reasonsBuilder strings.Builder
	var journeys []string

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

		highlight := containsIntermediateStation(service, departureStationCRS, arrivalStationCRS)
		if highlight {
			color.New(color.BgBlue).Fprint(&builder, row)
			statusColor.Fprintf(&builder, "%-10s %-20s %-40s\n", status, eta, service.Operator)
			journeys = append(journeys, formatJourney(service))
		} else {
			builder.WriteString(row)
			statusColor.Fprintf(&builder, "%-10s %-20s %-40s\n", status, eta, service.Operator)
		}

		if service.DelayReason != nil {
			reasonsBuilder.WriteString(fmt.Sprintf("\t‣ %s to %s - %s\n", origin, arrivalStationName, *service.DelayReason))
		}
	}

	builder.WriteString("=========================================================================================================\n")
	if reasonsBuilder.Len() > 0 {
		builder.WriteString("Reasons for delays/cancellations:\n")
		builder.WriteString(reasonsBuilder.String())
		builder.WriteString("=========================================================================================================\n")
	}

	if showJourneys && len(journeys) > 0 {
		builder.WriteString("Highlighted Journeys:\n")
		for _, journey := range journeys {
			builder.WriteString(journey)
		}
		builder.WriteString("=========================================================================================================\n")
	}

	return builder.String()
}

func formatJourney(service *nr.TrainService) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("\n🚆 %s Journey ‣ %s to %s: \n",
		service.STD,
		service.Origin.Name,
		service.Destination.Name,
	))
	builder.WriteString(fmt.Sprintf("  ┌─── [Origin: %s]\n", service.Origin.Name))

	// Previous calling points (already visited)
	for _, callingPoint := range service.PreviousCallingPoints {
		arrivalTime := "N/A"
		if callingPoint.At != nil {
			arrivalTime = *callingPoint.At
		}
		builder.WriteString(color.BlueString(fmt.Sprintf("  │   %s - %s\n", callingPoint.Name, arrivalTime)))
	}

	// Subsequent calling points (yet to visit)
	for _, callingPoint := range service.SubsequentCallingPoints {
		estimatedTime := "N/A"
		if callingPoint.Et != nil {
			estimatedTime = *callingPoint.Et
		}
		builder.WriteString(fmt.Sprintf("  │   %s - %s\n", callingPoint.Name, estimatedTime))
	}

	builder.WriteString(fmt.Sprintf("  └─── [Destination: %s]\n", service.Destination.Name))
	return builder.String()
}

func containsIntermediateStation(service *nr.TrainService, departureStationCRS, destinationStationCRS string) bool {
	for _, callingPoint := range service.SubsequentCallingPoints {
		if callingPoint.CRS == departureStationCRS || callingPoint.CRS == destinationStationCRS {
			return true
		}
	}
	return false
}

func getStationName(crs string) string {
	if name, exists := nr.StationCodeToNameMap[nr.CRSCode(crs)]; exists {
		return name
	}
	return crs
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
