package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/njason/shouldwater"
	"github.com/njason/tomorrowio-api-client"
	"github.com/njason/weather-api-client"
)

type Config struct {
	TomorrowioApiKey string `yaml:"tomorrowioApiKey"`
	WeatherApiKey string `yaml:"weatherApiKey"`
	MailChimp        struct {
		ApiKey     string `yaml:"apiKey"`
		TemplateId uint   `yaml:"templateId"`
		ListId     string `yaml:"listId"`
	}
	Lat         float64 `yaml:"lat"`
	Lng         float64 `yaml:"lng"`
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}

	var historicalRecords []shouldwater.WeatherRecord
	for i := 0; i < 7; i++ {
		historicalRequest := weatherapi.NewHistoryRequest(config.Lat, config.Lng, time.Now().AddDate(0, 0, -i))
		historicalRecordsRaw, err := weatherapi.DoHistoryRequest(config.WeatherApiKey, historicalRequest)
		if err != nil {
			log.Fatalln(err.Error())
		}

		for _, record := range historicalRecordsRaw.Forecast.Forecastday[0].Hour {
			timestamp, err := time.Parse("2006-01-02 15:04", record.Time)
			if err != nil {
				log.Fatalln(err.Error())
			}
			historicalRecords = append(historicalRecords, shouldwater.WeatherRecord{
				Timestamp: timestamp,
				Temperature: record.TempC,
				Humidity: float64(record.Humidity),
				WindSpeed: record.WindKph,
				Precipitation: record.PrecipMm,
			})
		}
	}

	if len(historicalRecords) < shouldwater.HoursInWeek {
		log.Fatalln(errors.New("need at least a week's worth of data to run"))
	} else if len(historicalRecords) > shouldwater.HoursInWeek {
		// trim to the last week of data
		historicalRecords = historicalRecords[len(historicalRecords)-shouldwater.HoursInWeek:]
	}

	forecastRequest := tomorrowio.NewTimelinesRequest(fmt.Sprintf("%f, %f", config.Lat, config.Lng), "metric", "1h", "nowPlus1h", "nowPlus5d")
	forecastRecordsRaw, err := tomorrowio.DoTimelinesRequest(config.TomorrowioApiKey, forecastRequest)
	if err != nil {
		log.Fatalln(err.Error())
	}

	var forecastRecords []shouldwater.WeatherRecord
	for _, record := range forecastRecordsRaw.Data.Timelines[0].Intervals {
		forecastRecords = append(forecastRecords, shouldwater.WeatherRecord{
			Timestamp: record.StartTime,
			Temperature: record.Values.Temperature,
			Humidity: record.Values.Humidity,
			WindSpeed: record.Values.WindSpeed,
			Precipitation: record.Values.PrecipitationIntensity,
		})
	}

	shouldWater, err := shouldwater.ShouldWater(historicalRecords, forecastRecords)
	if err != nil {
		log.Fatalln(err.Error())
	}

	if shouldWater {
		err = createAndSendCampaign(config.MailChimp.ApiKey, config.MailChimp.TemplateId, config.MailChimp.ListId)
		if err != nil {
			log.Fatalln(err.Error())
		}
	} else {
		log.Println("Should not water this week")
	}
}

func loadConfig() (Config, error) {
	f, err := os.Open("config.yaml")
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var config Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func archiveRecordsFile(recordsFile string) error {
	fileName := strings.TrimSuffix(recordsFile, filepath.Ext(recordsFile))
	archiveFile := fmt.Sprintf("%s_archive.csv", fileName)
	err := os.Rename(recordsFile, archiveFile)

	if err != nil {
		return err
	}

	return nil
}
