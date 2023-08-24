package main

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"
	"github.com/njason/shouldwater"
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

const FiveGallonBucket = 18.93  // in liters

func main() {
	var testMode = flag.Bool("test", false, "Will not send notifications")
	flag.Parse()

	config, err := loadConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}

	historicalRecords, err := getHistoricalRecords(config.WeatherApiKey, config.Lat, config.Lng)
	if err != nil {
		log.Fatalln(err.Error())
	}

	forecastRecords, err := getForecastRecords(config.TomorrowioApiKey, config.Lat, config.Lng)
	if err != nil {
		log.Fatalln(err.Error())
	}

	amountToWater, err := shouldwater.ShouldWater(historicalRecords, forecastRecords)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// less than a quarter of a bucket is not worth watering
	if amountToWater >= FiveGallonBucket / 4 {
		log.Printf("Should water %f buckets", amountToWater / FiveGallonBucket)

		if !*testMode {
			err = createAndSendCampaign(config.MailChimp.ApiKey, config.MailChimp.TemplateId, config.MailChimp.ListId)
			if err != nil {
				log.Fatalln(err.Error())
			}
		}
	} else {
		log.Println("Should not water")
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
