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
	gh "github.com/ragurney/term-check/pkg/github"
	"github.com/ragurney/term-check/pkg/lib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/waigani/diffparser"
)

const (
	checkName               = "Term Check"
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
		"created":     {},
		"rerequested": {},
	}
	pullRequestRelevantActions = map[string]struct{}{
		"opened":   {},
		"reopened": {},
	}
	flaggedTerms = []string{
		"master",
		"slave",
	}
)

// Bot is a type containing config for the GitHub bot logic
type Bot struct {
	client           *gh.Client
	server           *gh.Server
	privateKeyPath   string
	webhookSecretKey string
	appID            int
}

// New creates a new instance of Bot, taking in BotOptions
func New(options ...Option) *Bot {
	zerolog.TimeFieldFormat = ""

	b := Bot{}

	for _, option := range options {
		option(&b)
	}

	b.client = gh.NewClient(
		gh.WithPrivateKeyPath(b.privateKeyPath),
		gh.WithAppID(b.appID),
	)

	b.server = gh.NewServer(
		gh.WithWebhookSecretKey(b.webhookSecretKey),
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
func (b *Bot) HandleEvent(event interface{}) { //TODO DRY
	switch event := event.(type) {
	case *github.CheckSuiteEvent:
		log.Info().Msg("CheckSuiteEvent received")

		i := event.GetInstallation()
		cs := event.GetCheckSuite()

		if id := cs.GetApp().GetID(); id != int64(b.appID) {
			log.Error().Msgf("Event App ID of %d does not match Bot's App ID", id)
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
			log.Error().Msgf("Event App ID of %d does not match Bot's App ID", id)
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
	annotations, err := createAnnotations(ctx, pr, r, ghc)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create annotations for SHA %s", headSHA)
		return
	}

	cro := github.CreateCheckRunOptions{
		Name:        checkName,
		HeadBranch:  pr.GetHead().GetRef(),
		HeadSHA:     headSHA,
		Status:      github.String("completed"),
		CompletedAt: &github.Timestamp{time.Now()},
		Output: &github.CheckRunOutput{
			Title:            github.String("Term Check"),
			Text:             github.String("Placeholder text"),
			AnnotationsCount: github.Int(len(annotations)),
			Annotations:      annotations,
		},
	}
	// presence of annotations signals there is usage of flagged terms
	if len(annotations) > 0 {
		cro.Conclusion = github.String(checkFailureConclusion)
		cro.Output.Summary = github.String("⚠️ Flagged terms found.")
	} else {
		cro.Conclusion = github.String(checkSuccessConclusion)
		cro.Output.Summary = github.String("✅ No flagged terms found.")
	}

	_, resp, err := ghc.Checks.CreateCheckRun(ctx, r.GetOwner().GetLogin(), r.GetName(), cro)
	if code := resp.StatusCode; err != nil || (code < 200 || code > 299) {
		log.Error().Err(err).Msgf("Failed to POST CheckRun for SHA %s", headSHA)
	} else {
		log.Debug().Msgf("Successfully created CheckRun for SHA %s", headSHA)
	}
}

func createAnnotations(ctx context.Context, pr *github.PullRequest, r *github.Repository, ghc *github.Client) ([]*github.CheckRunAnnotation, error) {
	headSHA := pr.GetHead().GetSHA()
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

	re, _ := regexp.Compile(strings.Join(flaggedTerms, "|"))
	var annotations = []*github.CheckRunAnnotation{}

	for _, f := range parsedDiff.Files {
		for _, h := range f.Hunks {
			adds := h.NewRange
			for _, l := range adds.Lines {
				if matches := lib.Unique(re.FindAllString(l.Content, -1)); len(matches) > 0 {
					annotations = append(annotations, createAnnotation(f, l, matches))
				}
			}
		}
	}

	return annotations, nil
}

func createAnnotation(f *diffparser.DiffFile, l *diffparser.DiffLine, m []string) (a *github.CheckRunAnnotation) {
	msg := fmt.Sprintf("Please consider changing the following terms on this line: `%s`", strings.Join(m, ", "))

	return &github.CheckRunAnnotation{
		Path:            github.String(f.NewName),
		StartLine:       github.Int(l.Number),
		EndLine:         github.Int(l.Number),
		AnnotationLevel: github.String(checkRunAnnotationLevel),
		Message:         github.String(msg),
		Title:           github.String("Term Notice"),
	}
}
