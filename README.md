# WaterDuty - The tree watering stewardship notification tool

A low cost solution for automating alerting tree stewards to water their nearby unestablished trees (less than two years old) and gardens during periods of low rainfall. This application uses forecasted and recent historical weather data to determine if watering is needed. This solutions uses [tomorrow.io](https://www.tomorrow.io/) and [weatherapi.com](https://www.weatherapi.com/) for weather data and [mailchimp](https://mailchimp.com/) for emailing alerts.

[Guide for watering unestablished trees](https://vimeo.com/416031708#t=5m35s).

## Setup

### Config

Copy the `config-template.yaml` into a new file `config.yaml`, and update the following fields:

- `tomorrowioApiKey`: After creating a free tomorrow.io account, find the `Secret Key` [here](https://app.tomorrow.io/development/keys).
- `weatherApiKey`: After creating a free weatherapi.com account, find the API Key [here](https://www.weatherapi.com/my/).
- `mailchimp.apiKey`: After creating a free mailchimp account, create an API key [here](https://admin.mailchimp.com/account/api/).
- `mailchimp.templateId`: [Create a template](https://mailchimp.com/help/create-a-template-with-the-template-builder/) to use for alerting tree stewards to water. [This](https://us13.admin.mailchimp.com/templates/share?id=174361973_a7f368481da096f6c0df_us13) is the template used in NYC you can use as a starting point.
- `mailchimp.listId`: Use [this guide](https://mailchimp.com/help/find-audience-id/) to find the list/audience ID.
- `lat`, `lng`: The coordinates of where to run. You can use [Google Maps](https://support.google.com/maps/answer/18539) to find coordinates, format `lat, lng`.

### Testing

To test without sending emails, use the `-test` flag while running.

### Deploying

The binary will need to be in the same directory as `config.yaml` to run.

The application should be scheduled to run once a week to decide if an alert goes out. It is recommended to run on Fridays so tree stewards can water over the weekends. The cron for this would be `0 16 * 4-10 FRI cd /path/to && ./waterduty 1>>waterduty.log 2>&1`. Note that the application should only be run during times of the year when trees are not dormant, from early spring to early fall.

To setup automatic scheduling on a server, for macOS and linux use [cron](https://phoenixnap.com/kb/set-up-cron-job-linux). For Windows use [Task Scheduler](https://www.windowscentral.com/how-create-automated-task-using-task-scheduler-windows-10).
