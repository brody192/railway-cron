package main

import (
	"main/internal/railway"
	"main/internal/schedule"
	"os"
	"time"

	"github.com/brody192/logger"

	"log/slog"

	"main/internal/retry"

	"github.com/go-co-op/gocron"
)

func main() {
	// get the account token from the environment, fail if missing
	railwayToken := os.Getenv("RAILWAY_ACCOUNT_TOKEN")
	if railwayToken == "" {
		logger.Stderr.Error("missing 'RAILWAY_ACCOUNT_TOKEN' environment variable")
		os.Exit(1)
	}

	// parse schedule configuration strings from the environment with the given prefix
	schedules, err := schedule.ParseFromEnv("SCHEDULE_")
	if err != nil {
		logger.Stderr.Error("error parsing schedules from environment", logger.ErrAttr(err))
		os.Exit(1)
	}

	if len(schedules) == 0 {
		logger.Stderr.Error("no schedules found")
		logger.Stdout.Info("set at least one or more schedules with 'SCHEDULE_1', 'SCHEDULE_2', etc. variables")
		os.Exit(1)
	}

	logger.Stdout.Info("parsed schedules from environment successfully",
		slog.Int("found_schedules", len(schedules)),
	)

	railwayClient := railway.NewAuthedClient(railwayToken)

	backoffParams := retry.GetBackoffParams()

	// print schedules for viewing purposes
	for i := range schedules {
		project, err := railway.Project(railwayClient, schedules[i].ProjectID)
		if err != nil {
			logger.Stderr.Error("failed retrieving project information", logger.ErrAttr(err), slog.String("project_id", schedules[i].ProjectID))
			os.Exit(1)
		}

		logger.Stdout.Info("found schedule",
			slog.String("service_id", schedules[i].ServiceID),
			slog.String("project_name", project.Project.Name),
			slog.String("action", string(schedules[i].Action)),
			slog.String("expression", schedules[i].Expression),
		)
	}

	logger.Stdout.Info("access to all projects defined in the schedule(s) confirmed")

	// cron job function that will be executed on the set schedules
	cronTask := func(jobDetails schedule.Schedule) {
		// get the friendly service name, looking at just the service id can get very confusing
		friendlyName, err := railwayClient.GetFriendlyName(jobDetails.ServiceID)
		if err != nil {
			logger.Stderr.Warn("error retrieving friendly service name", logger.ErrAttr(err))
		}

		// default slog attributes
		slogAttr := []any{
			slog.String("service_name", friendlyName),
			slog.String("action", string(jobDetails.Action)),
			slog.String("expression", jobDetails.Expression),
			slog.String("service_id", jobDetails.ServiceID),
		}

		logger.Stdout.Info("starting cron job", slogAttr...)

		var latestDeploymentID string
		err = retry.RetryWithBackoff(
			func() error {
				var err error
				latestDeploymentID, err = railwayClient.GetLatestDeploymentID(jobDetails)
				return err
			},
			backoffParams,
		)

		if err != nil {
			slogAttr = append(slogAttr, logger.ErrAttr(err))
			logger.Stderr.Error("error getting latest deployment for given service after retries", slogAttr...)
			return
		}

		// run action depending on the action type
		switch jobDetails.Action {
		case schedule.ActionRedeploy:
			retry.RetryWithBackoff(
				func() error {
					_, err = railway.DeploymentRedeploy(railwayClient, latestDeploymentID)
					return err
				},
				backoffParams,
			)
			if err != nil {
				slogAttr = append(slogAttr, logger.ErrAttr(err))
				logger.StderrWithSource.Error("error redeploying the given service", slogAttr...)
				return
			}
		case schedule.ActionRestart:
			retry.RetryWithBackoff(
				func() error {
					_, err = railway.DeploymentRestart(railwayClient, latestDeploymentID)
					return err
				},
				backoffParams,
			)
			if err != nil {
				slogAttr = append(slogAttr, logger.ErrAttr(err))
				logger.StderrWithSource.Error("error restarting the given service", slogAttr...)
				return
			}
		default:
			slogAttr = append(slogAttr, slog.String("action", string(jobDetails.Action)))
			logger.StderrWithSource.Error("invalid action", slogAttr...)
			return
		}

		logger.Stdout.Info("cron job completed successfully", slogAttr...)
	}

	// create a new cron schedular in utc time
	scheduler := gocron.NewScheduler(time.UTC)

	// register all scheduled jobs
	for _, job := range schedules {
		_, err := scheduler.Cron(job.Expression).Do(cronTask, job)
		if err != nil {

			logger.StderrWithSource.Error("error registering schedule with cron", logger.ErrAttr(err))
			return
		}
	}

	scheduledJobs, registeredJobs := len(schedules), len(scheduler.Jobs())

	// check if we registered the same amount of jobs as there was schedules
	if scheduledJobs != registeredJobs {
		logger.StderrWithSource.Warn("cron jobs registered not equal to number of parsed schedules",
			slog.Int("scheduled_jobs", scheduledJobs),
			slog.Int("registered_jobs", registeredJobs),
		)
		os.Exit(1)
	}

	logger.Stdout.Info("starting scheduler", slog.Int("registered_jobs", registeredJobs))

	// start the scheduler in blocking mode
	scheduler.StartBlocking()
}
