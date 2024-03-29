# Term Check [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/zendesk/term-check?tab=doc) [![GolangCI](https://golangci.com/badges/github.com/golangci/golangci-lint.svg)](https://golangci.com)

This bot is for our Inclusive Language initiative inside Zendesk Engineering.

![Screen Shot 2020-08-19 at 11 00 23 AM](https://user-images.githubusercontent.com/15261525/90672683-582bbb00-e20b-11ea-844e-3ddc2ab85c29.png)

## Configuration

### Bot Configuration

Configuration for the bot's behavior is contained in `config.yaml`, e.x.

```yaml
# Any shared configuration between fields
shared:
  # ID of the GitHub application
  appID: &appID 123456
botConfig:
  appID: *appID
  # List of terms to look for and flag in code
  termList:
    - slave
  # Name of the check. Will appear in the status list and as the title on the 'details' page
  checkName: Inclusive Language Check
  # Check summary to set when no terms are found
  checkSuccessSummary: Looks good! 😇
  # Check summary to set when terms are found
  checkFailureSummary: 👋 exclusive language
  # Generic check details text
  checkDetails: "Language check results:"
  # Text for the title of check annotations created for each flagged term in the code
  annotationTitle: Exclusive Language
  # Text for the body of each annotation. Supports one format string [%s] which will be replaced by the flagged terms
  # found on that line
  annotationBody: |
    Hi there! 👋 I see you used the term(s) [%s] here. This language is exclusionary for members of our community,
    please consider changing it.
clientConfig:
  appID: *appID
  # Path to the private key generated for the GitHub application
  privateKeyPath: /secrets/PRIVATE_KEY
```

### Repo-Specific Configuration

Certain behaviors are configurable on a per repository basis. Add a `.github/term-check.yaml` file to your
repository based off of the following template:

```yaml
# An array of patterns following .gitignore rules (http://git-scm.com/docs/gitignore) specifying which files and
# directories should be ignored by the app
ignore:
  - foo
  - bar/
```

## Deploying Your Own Instance
See [docs/deploy.md](docs/deploy.md) for instructions to deploy your own term-check instance.

## Copyright and License

Copyright 2019 Zendesk, Inc.
Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0 Unless required by applicable law or
agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and limitations under the License.
