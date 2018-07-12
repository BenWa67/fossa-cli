package test

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/fossas/fossa-cli/api"
	"github.com/fossas/fossa-cli/api/fossa"
	"github.com/fossas/fossa-cli/cmd/fossa/cmdutil"
	"github.com/fossas/fossa-cli/cmd/fossa/flags"
	"github.com/fossas/fossa-cli/config"
	"github.com/fossas/fossa-cli/log"
)

var Timeout = "timeout"

const pollRequestDelay = 8 * time.Second
const buildsEndpoint = "/api/revisions/%s/build"
const revisionsEndpoint = "/api/revisions/%s"

var Cmd = cli.Command{
	Name:   "test",
	Usage:  "Test current revision against FOSSA scan status and exit with errors if issues are found",
	Action: Run,
	Flags: flags.WithGlobalFlags(flags.WithAPIFlags(([]cli.Flag{
		cli.IntFlag{Name: Timeout, Value: 10 * 60, Usage: "duration to wait for build completion (in seconds)"},
	}))),
}

var _ cli.ActionFunc = Run

func Run(ctx *cli.Context) error {
	err := cmdutil.InitWithAPI(ctx)
	if err != nil {
		log.Logger.Fatalf("Could not initialize: %s", err.Error())
	}

	issues, err := Do(time.After(time.Duration(ctx.Int(Timeout)) * time.Second))
	if err != nil {
		log.Logger.Fatalf("Could not test revision: %s", err.Error())
	}

	log.Logger.Debugf("Test succeeded: %#v", issues)
	if len(issues) == 0 {
		fmt.Fprintln(os.Stderr, "Test passed! 0 issues found")
		return nil
	}

	pluralizedIssues := "issues"
	if len(issues) == 1 {
		pluralizedIssues = "issue"
	}
	fmt.Fprintf(os.Stderr, "Test failed! %d %s found\n", len(issues), pluralizedIssues)

	marshalled, err := json.Marshal(issues)
	if err != nil {
		log.Logger.Fatalf("Could not marshal unresolved issues: %s", err)
	}
	fmt.Println(string(marshalled))

	os.Exit(1)
	return nil
}

func Do(stop <-chan time.Time) ([]fossa.Issue, error) {
	log.ShowSpinner("Waiting for analysis to complete...")

	project := fossa.Locator{
		Fetcher:  config.Fetcher(),
		Project:  config.Project(),
		Revision: config.Revision(),
	}

	_, err := CheckBuild(project, stop)
	if err != nil {
		log.Logger.Fatalf("Could not load build: %s", err.Error())
	}

	log.ShowSpinner("Waiting for FOSSA scan results...")

	issues, err := CheckIssues(project, stop)
	if err != nil {
		log.Logger.Fatalf("Could not load issues: %s", err.Error())
	}

	log.StopSpinner()
	return issues, nil
}

func CheckBuild(locator fossa.Locator, stop <-chan time.Time) (fossa.Build, error) {
	for {
		select {
		case <-stop:
			return fossa.Build{}, errors.New("timed out while waiting for build")
		default:
			build, err := fossa.GetBuild(locator)
			if _, ok := err.(api.TimeoutError); ok {
				time.Sleep(pollRequestDelay)
				continue
			}
			if err != nil {
				return fossa.Build{}, errors.Wrap(err, "error while loading build")
			}
			switch build.Task.Status {
			case "SUCCEEDED":
				return build, nil
			case "FAILED":
				return fossa.Build{}, fmt.Errorf("failed to analyze build #%d: %s (visit FOSSA or contact support@fossa.io)", build.ID, build.Error)
			default:
				time.Sleep(pollRequestDelay)
			}
		}
	}
}

func CheckIssues(locator fossa.Locator, stop <-chan time.Time) ([]fossa.Issue, error) {
	for {
		select {
		case <-stop:
			return nil, errors.New("timed out while waiting for scan")
		default:
			issues, err := fossa.GetIssues(locator)
			if _, ok := err.(api.TimeoutError); ok {
				time.Sleep(pollRequestDelay)
				continue
			}
			if err != nil {
				return nil, errors.Wrap(err, "error while loading issues")
			}
			log.Logger.Debugf("Got issues: %#v", issues)
			return issues, nil
		}
	}
}