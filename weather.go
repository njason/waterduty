package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/njason/shouldwater"
	"github.com/njason/tomorrowio-api-client"
	"github.com/njason/weather-api-client"
)

func getHistoricalRecords(weatherApiKey string, lat float64, lng float64) ([]shouldwater.WeatherRecord, error) {
	var historicalRecords []shouldwater.WeatherRecord

	now := time.Now().UTC()
	weekAgo := now.AddDate(0, 0, -7)

	totalPrecipitation := .0

	for i := 0; i < 8; i++ {
		queryTime := now.AddDate(0, 0, -i)
		historicalRequest := weatherapi.NewHistoryRequest(lat, lng, queryTime)
		historicalRecordsRaw, err := weatherapi.DoHistoryRequest(weatherApiKey, historicalRequest)
		if err != nil {
			return []shouldwater.WeatherRecord{}, err
		}

		for _, record := range historicalRecordsRaw.Forecast.Forecastday[0].Hours {
			timestamp, err := time.Parse("2006-01-02 15:04", record.Time)
			if err != nil {
				return []shouldwater.WeatherRecord{}, err
			}

			if now.After(timestamp) && weekAgo.Before(timestamp) {
				historicalRecords = append(historicalRecords, shouldwater.WeatherRecord{
					Timestamp:     timestamp,
					Temperature:   record.TempC,
					Humidity:      float64(record.Humidity),
					WindSpeed:     record.WindKph,
					Precipitation: record.PrecipMm,
				})
				totalPrecipitation += record.PrecipMm
			}
		}
	}

	if len(historicalRecords) < shouldwater.HoursInWeek {
		return historicalRecords, errors.New("need at least a week's worth of data to run")
	} else if len(historicalRecords) > shouldwater.HoursInWeek {
		// trim to the last week of data
		historicalRecords = historicalRecords[len(historicalRecords)-shouldwater.HoursInWeek:]
	}

	log.Printf("Historical precipitation: %.2f inches\n", totalPrecipitation/25.4)

	return historicalRecords, nil
}

func getForecastRecords(tomorrowioApiKey string, lat float64, lng float64) ([]shouldwater.WeatherRecord, error) {
	forecastRequest := tomorrowio.NewTimelinesRequest(fmt.Sprintf("%f, %f", lat, lng), "metric", "1h", "nowPlus1h", "nowPlus5d")
	forecastRecordsRaw, err := tomorrowio.DoTimelinesRequest(tomorrowioApiKey, forecastRequest)
	if err != nil {
		return []shouldwater.WeatherRecord{}, err
	}

	totalPrecipitation := .0

	var forecastRecords []shouldwater.WeatherRecord
	for _, record := range forecastRecordsRaw.Data.Timelines[0].Intervals {
		forecastRecords = append(forecastRecords, shouldwater.WeatherRecord{
			Timestamp:     record.StartTime,
			Temperature:   record.Values.Temperature,
			Humidity:      record.Values.Humidity,
			WindSpeed:     record.Values.WindSpeed,
			Precipitation: record.Values.PrecipitationIntensity,
		})
		totalPrecipitation += record.Values.PrecipitationIntensity
	}

	log.Printf("Forecast precipitation: %.2f inches\n", totalPrecipitation/25.4)

	return forecastRecords, nil
}
