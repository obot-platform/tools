package main

import (
	"fmt"
	"github.com/icodealot/noaa"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: weather <command>")
	}

	command := os.Args[1]

	latitude := os.Getenv("LATITUDE")
	if latitude == "" {
		log.Fatal("required latitude parameter not set")
	}
	longitude := os.Getenv("LONGITUDE")
	if longitude == "" {
		log.Fatal("required longitude parameter not set")
	}

	switch command {
	case "getWeeklyForecast":
		forecast, err := noaa.Forecast(latitude, longitude)
		if err != nil {
			log.Fatalf("error getting the forecast: %v", err)
			return
		}
		for _, period := range forecast.Periods {
			fmt.Printf("%-20s ---> %.0f%s\n", period.Name, period.Temperature, period.TemperatureUnit)
		}
	case "getHourlyForecast":
		forecast, err := noaa.HourlyForecast(latitude, longitude)
		if err != nil {
			log.Fatalf("error getting the forecast: %v", err)
		}
		for _, period := range forecast.Periods {
			fmt.Printf("%-20s ---> %.0f%s\n", period.StartTime, period.Temperature, period.TemperatureUnit)
		}
	}

}
