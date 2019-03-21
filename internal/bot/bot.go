// Package bot contains all of the main logic for handling GitHub events, including creating CheckRuns for each
// Pull Request
package bot

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v18/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sabhiram/go-gitignore"
	"github.com/waigani/diffparser"
	"github.com/zendesk/term-check/internal/config"
	gh "github.com/zendesk/term-check/pkg/github"
	"github.com/zendesk/term-check/pkg/lib"
)

const (
	checkSuccessConclusion  = "success"
	checkFailureConclusion  = "neutral"
	checkRunAnnotationLevel = "warning"
)

var (
	checkSuiteRelevantActions = map[string]struct{}{
		"requested":   {},
		"rerequested": {},
	}
	checkRunRelevantActions = map[string]struct{}{
		"rerequested": {},
	}
	pullRequestRelevantActions = map[string]struct{}{
		"opened":   {},
		"reopened": {},
	}
)

// Bot is a type containing config for the GitHub bot logic
type Bot struct {
	client              *gh.Client
	server              *gh.Server
	privateKeyPath      string
	webhookSecretKey    string
	appID               int
	termList            []string
	checkName           string
	checkSuccessSummary string
	checkFailureSummary string
	checkDetails        string
	annotationTitle     string
	annotationBody      string
}

// New creates a new instance of Bot, taking in BotOptions
func New(botConfig *config.BotConfig, clientConfig *config.ClientConfig, serverConfig *config.ServerConfig) *Bot {
	zerolog.TimeFieldFormat = ""

	b := Bot{
		appID:               botConfig.AppID,
		termList:            botConfig.TermList,
		checkName:           botConfig.CheckName,
		checkSuccessSummary: botConfig.CheckSuccessSummary,
		checkFailureSummary: botConfig.CheckFailureSummary,
		checkDetails:        botConfig.CheckDetails,
		annotationTitle:     botConfig.AnnotationTitle,
		annotationBody:      botConfig.AnnotationBody,
	}

	b.client = gh.NewClient(
		gh.WithPrivateKeyPath(clientConfig.PrivateKeyPath),
		gh.WithAppID(clientConfig.AppID),
	)

	b.server = gh.NewServer(
		gh.WithWebhookSecretKey(serverConfig.WebhookSecretKey),
		gh.WithEventHandler(&b),
	)

	return &b
}

// Start starts the bot server
func (b *Bot) Start() {
	log.Debug().Msg("Starting bot...")

	b.server.Start()
}

// HandleEvent interface implementation for Server to pass incoming GitHub events to
func (b *Bot) HandleEvent(event interface{}) {
	switch event := event.(type) {
	case *github.CheckSuiteEvent:
		log.Info().Msg("CheckSuiteEvent received")

		i := event.GetInstallation()
		cs := event.GetCheckSuite()

		if id := cs.GetApp().GetID(); id != int64(b.appID) {
			log.Error().Msgf("Event App ID of %d does not match Bot's App ID of %d", id, b.appID)
			return
		}
		if action := event.GetAction(); !lib.Contains(checkSuiteRelevantActions, action) {
			log.Debug().Msgf("Unhandled action received: %s. Discarding...", action)
			return
		}

		r := event.GetRepo()
		gClient := b.client.CreateClient(int(i.GetID())) // truncating
		ctx := context.Background()

		for _, pr := range cs.PullRequests {
			b.createCheckRun(ctx, pr, r, gClient)
		}
	case *github.CheckRunEvent:
		log.Debug().Msg("CheckRunEvent received")

		i := event.GetInstallation()
		cr := event.GetCheckRun()

		if id := cr.GetApp().GetID(); id != int64(b.appID) {
			log.Error().Msgf("Event App ID of %d does not match Bot's App ID of %d", id, b.appID)
			return
		}
		if action := event.GetAction(); !lib.Contains(checkRunRelevantActions, action) {
			log.Debug().Msgf("Unhandled action received: %s. Discarding...", action)
			return
		}

		r := event.GetRepo()
		gClient := b.client.CreateClient(int(i.GetID())) // truncating
		ctx := context.Background()

		for _, pr := range cr.PullRequests {
			b.createCheckRun(ctx, pr, r, gClient)
		}
	case *github.PullRequestEvent:
		log.Debug().Msg("PullRequestEvent received")

		i := event.GetInstallation()

		if action := event.GetAction(); !lib.Contains(pullRequestRelevantActions, action) {
			log.Debug().Msgf("Unhandled action received: %s. Discarding...", action)
			return
		}

		gClient := b.client.CreateClient(int(i.GetID())) // truncating
		ctx := context.Background()

		b.createCheckRun(ctx, event.GetPullRequest(), event.GetRepo(), gClient)
	default:
		log.Debug().Msgf("Unhandled event received: %s. Discarding...", reflect.TypeOf(event).Elem().Name())
	}

	return
}

func (b *Bot) createCheckRun(ctx context.Context, pr *github.PullRequest, r *github.Repository, ghc *github.Client) {
	headSHA := pr.GetHead().GetSHA()

	log.Debug().Msgf("Creating CheckRun for SHA %s...", headSHA)
	annotations, err := b.createAnnotations(ctx, pr, r, ghc)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create annotations for SHA %s", headSHA)
		return
	}

	cro := github.CreateCheckRunOptions{
		Name:        b.checkName,
		HeadBranch:  pr.GetHead().GetRef(),
		HeadSHA:     headSHA,
		Status:      github.String("completed"),
		CompletedAt: &github.Timestamp{time.Now()},
		Output: &github.CheckRunOutput{
			Title:            github.String(b.checkName),
			Text:             github.String(b.checkDetails),
			AnnotationsCount: github.Int(len(annotations)),
			Annotations:      annotations,
		},
	}
	// presence of annotations signals there is usage of flagged terms
	if len(annotations) > 0 {
		cro.Conclusion = github.String(checkFailureConclusion)
		cro.Output.Summary = github.String(b.checkFailureSummary)
	} else {
		cro.Conclusion = github.String(checkSuccessConclusion)
		cro.Output.Summary = github.String(b.checkSuccessSummary)
	}

	_, resp, err := ghc.Checks.CreateCheckRun(ctx, r.GetOwner().GetLogin(), r.GetName(), cro)
	if code := resp.StatusCode; err != nil || (code < 200 || code > 299) {
		log.Error().Err(err).Msgf("Failed to POST CheckRun for SHA %s", headSHA)
	} else {
		log.Debug().Msgf("Successfully created CheckRun for SHA %s", headSHA)
	}
}

func (b *Bot) createAnnotations(ctx context.Context, pr *github.PullRequest, r *github.Repository, ghc *github.Client) ([]*github.CheckRunAnnotation, error) {
	headSHA := pr.GetHead().GetSHA()

	// Get PR diff
	diff, resp, err := ghc.PullRequests.GetRaw( // TODO: refactor to move methods making requests to Client?
		ctx,
		r.GetOwner().GetLogin(),
		r.GetName(),
		pr.GetNumber(),
		github.RawOptions{Type: github.Diff},
	)
	if err != nil || resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("Failed to get diff for %s: %s", headSHA, err)
		return []*github.CheckRunAnnotation{}, e
	}
	parsedDiff, err := diffparser.Parse(diff)
	if err != nil {
		e := fmt.Errorf("Failed to parse diff for %s: %s", headSHA, err)
		return []*github.CheckRunAnnotation{}, e
	}

	re, _ := regexp.Compile(strings.Join(b.termList, "|"))
	var annotations = []*github.CheckRunAnnotation{}

	for _, f := range parsedDiff.Files {
		// Skip over any files listed in `ignore`
		if ignoredByRepo(ctx, r, headSHA, f.NewName, ghc) {
			continue
		}

		for _, h := range f.Hunks {
			adds := h.NewRange
			for _, l := range adds.Lines {
				if matches := lib.Unique(re.FindAllString(l.Content, -1)); len(matches) > 0 {
					annotations = append(annotations, b.createAnnotation(f, l, matches))
				}
			}
		}
	}

	return annotations, nil
}

func (b *Bot) createAnnotation(f *diffparser.DiffFile, l *diffparser.DiffLine, m []string) (a *github.CheckRunAnnotation) {
	msg := fmt.Sprintf(b.annotationBody, strings.Join(m, ", ")) // Expects %s format string in body
	msg = strings.Split(msg, "%!")[0]                           // Remove formatting error if user doesn't provide format string in body

	return &github.CheckRunAnnotation{
		Path:            github.String(f.NewName),
		StartLine:       github.Int(l.Number),
		EndLine:         github.Int(l.Number),
		AnnotationLevel: github.String(checkRunAnnotationLevel),
		Message:         github.String(msg),
		Title:           github.String(b.annotationTitle),
	}
}

func ignoredByRepo(ctx context.Context, r *github.Repository, headSHA, filename string, ghc *github.Client) bool {
	ic := config.GetRepoConfig(ctx, r, headSHA, ghc) // Get repository configuration

	if ignorePatterns := ic.Ignore; ignorePatterns != nil {
		ignoreMatcher, err := ignore.CompileIgnoreLines(ignorePatterns...)

		if err != nil {
			log.Warn().Err(err).Msg("Disregarding `ignore` configuration")
			return false
		}
		return ignoreMatcher.MatchesPath(filename)
	}

	return false
}
