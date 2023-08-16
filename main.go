package main

import (
	"main/internal/logger"
	"main/internal/railway"
	"main/internal/schedule"
	"main/internal/tools"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"golang.org/x/exp/slog"
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
		logger.Stderr.Error("error parsing schedules from environment: " + tools.ErrStr(err))
		os.Exit(1)
	}

	logger.Stdout.Info("parsed schedules from environment successfully",
		slog.Int("found_schedules", len(schedules)),
	)

	// print schedules for viewing purposes
	for _, schedule := range schedules {
		logger.Stdout.Info("found schedule",
			slog.String("service_id", schedule.ServiceID),
			slog.String("action", string(schedule.Action)),
			slog.String("expression", schedule.Expression),
		)
	}

	railwayClient := railway.NewAuthedClient(railwayToken)

	// check if we have account level authorization, if not, fail early
	meResp, err := railway.Me(railwayClient)
	if err != nil {
		logger.Stderr.Error("failed retrieving user information: " + tools.ErrStr(err))
		os.Exit(1)
	}

	logger.Stdout.Info("user information retrieved successfully", slog.String("username", meResp.Me.Name), slog.String("email", meResp.Me.Email))

	// cron job function that will be executed on the set schedules
	cronTask := func(jobDetails schedule.Schedule) {
		// get the friendly service name, looking at just the service id can get very confusing
		friendlyName, err := railwayClient.GetFriendlyName(jobDetails.ServiceID)
		if err != nil {
			logger.Stderr.Warn("error retrieving friendly service name: " + tools.ErrStr(err))
		}

		// default slog attributes
		slogAttr := []any{
			slog.String("service_name", friendlyName),
			slog.String("action", string(jobDetails.Action)),
			slog.String("expression", jobDetails.Expression),
			slog.String("service_id", jobDetails.ServiceID),
		}

		logger.Stdout.Info("starting cron job", slogAttr...)

		// retrieve latest active or complete deployment from service
		latestDeploymentID, err := railwayClient.GetLatestDeploymentID(jobDetails)
		if err != nil {
			logger.Stderr.Error(tools.ErrStr(err), slogAttr...)
			return
		}

		// run action depending on the action type
		switch jobDetails.Action {
		case schedule.ActionRedeploy:
			_, err = railway.DeploymentRedeploy(railwayClient, latestDeploymentID)
			if err != nil {
				logger.StderrWithSource.Error(tools.ErrStr(err), slogAttr...)
				return
			}
		case schedule.ActionRestart:
			_, err = railway.DeploymentRestart(railwayClient, latestDeploymentID)
			if err != nil {
				logger.StderrWithSource.Error(tools.ErrStr(err), slogAttr...)
				return
			}
		default:
			logger.StderrWithSource.Error("invalid action: "+string(jobDetails.Action), slogAttr...)
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
			logger.StderrWithSource.Error(tools.ErrStr(err))
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
