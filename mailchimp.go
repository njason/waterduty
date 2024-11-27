package main

import (
	"time"

	"github.com/njason/mailchimp_client"
)

func createAndSendCampaign(apiKey string, templateId uint, listId string) error {
	client := mailchimp_client.New(apiKey)
	client.Timeout = 60 * time.Second

	createCampaignRequest := mailchimp_client.CampaignCreationRequest{
		Type: mailchimp_client.CAMPAIGN_TYPE_REGULAR,
		Recipients: mailchimp_client.CampaignCreationRecipients{
			ListId: listId,
		},
		Settings: mailchimp_client.CampaignCreationSettings{
			SubjectLine: "It's time to water the trees!",
			Title:       "NYC unestablished tree watering alert",
			FromName:    "Water Duty",
			ReplyTo:     "noreply@waterduty.org",
			ToName:      "NYC Tree Stewards",
			TemplateId:  templateId,
		},
	}

	createCampaignResponse, err := client.CreateCampaign(&createCampaignRequest)
	if err != nil {
		return err
	}

	if createCampaignResponse == nil {
		return err
	}

	_, err = client.SendCampaign(createCampaignResponse.ID, nil)
	if err != nil {
		return err
	}
	return nil
}
