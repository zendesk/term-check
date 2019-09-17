# Deployment

This repository is currently configured for deployment of the Zendesk term-check
bot. To deploy your own instance:

## Create a GitHub App

Your deployment will need a corresponding [GitHub App](https://developer.github.com/apps/).

1. [Create](https://github.com/settings/apps/new) your new GitHub App.
   - Basic info:
     - Add a name for your application.
     - Set "Homepage URL" to the cloned repository of the app.
     - Leave "Webhook URL" blank for now.
     - Generate a "Webhook Secret" with `openssl rand -base64 32` and save it for later use in deployment.
   - Permissions
     - Your app will need the following repository permissions:
       1. **Checks**: Read & write
       1. **Contents**: Read-only
       1. **Metadata**: Read-only
       1. **Pull requests**: Read & write
     - It will also need the following event subscriptions:
       1. Check run
       1. Pull request
1. Download the private key of the application.
1. Install the app on whichever repositories you want.

## Deploy Your App

1. Change the [config.yaml](../config.yaml) file to match your own app's configuration and preferences.
   - Set `appID` to the one given by your newly [created app](https://github.com/settings/apps).
   - Set `privateKeyPath` to be the path to the downloaded private key when your app is deployed.
1. Populate secret values
   - The bot expects the secret values `PRIVATE_KEY` and `WEBHOOK_SECRET_KEY` to be in files in a `secrets/<Secret Name>`, where each file contains the file name's corresponding value.
1. Deploy the app on a platform of your choice. This repo contains configuration files for a GCB and Kubernetes deployment process, but they would have to be tweaked for your own purposes. Once the application is deployed, update the GitHub App's "Webhook URL" to point to the url of your deployment.
