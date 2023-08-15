package main

import (
	"main/logger"
	"main/railway"
	"main/schedule"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"golang.org/x/exp/slog"
)

func main() {
	railwayToken := os.Getenv("RAILWAY_ACCOUNT_TOKEN")
	if railwayToken == "" {
		logger.Stderr.Error("missing 'RAILWAY_ACCOUNT_TOKEN' environment variable")
		os.Exit(1)
	}

	schedules, err := schedule.ParseFromEnv("SCHEDULE_")
	if err != nil {
		logger.Stderr.Error("error parsing schedules from environment: " + err.Error())
		os.Exit(1)
	}

	logger.Stdout.Info("parsed schedules from environment successfully",
		slog.Int("found_schedules", len(schedules)),
	)

	for _, schedule := range schedules {
		logger.Stdout.Info("found schedule",
			slog.String("service_id", schedule.ServiceID),
			slog.String("action", string(schedule.Action)),
			slog.String("expression", schedule.Expression),
		)
	}

	railwayClient := railway.NewAuthedClient(railwayToken)

	cronTask := func(jobDetails schedule.Schedule) {
		friendlyName, err := railwayClient.GetFriendlyName(jobDetails.ServiceID)
		if err != nil {
			logger.Stderr.Warn("error retrieving friendly service name: " + err.Error())
		}

		slogAttr := []any{
			slog.String("service_name", friendlyName),
			slog.String("action", string(jobDetails.Action)),
			slog.String("expression", jobDetails.Expression),
			slog.String("service_id", jobDetails.ServiceID),
		}

		logger.Stdout.Info("starting cron job", slogAttr...)

		latestDeploymentID, err := railwayClient.GetLatestDeploymentID(jobDetails)
		if err != nil {
			logger.Stderr.Warn(err.Error(), slogAttr...)
			return
		}

		switch jobDetails.Action {
		case schedule.ActionRedeploy:
			_, err = railway.DeploymentRedeploy(railwayClient, latestDeploymentID)
			if err != nil {
				logger.StderrWithSource.Error(err.Error(), slogAttr...)
				return
			}
		case schedule.ActionRestart:
			_, err = railway.DeploymentRestart(railwayClient, latestDeploymentID)
			if err != nil {
				logger.StderrWithSource.Error(err.Error(), slogAttr...)
				return
			}
		default:
			logger.StderrWithSource.Error("invalid action: "+string(jobDetails.Action), slogAttr...)
			return
		}

		logger.Stdout.Info("cron job completed successfully", slogAttr...)
	}

	scheduler := gocron.NewScheduler(time.UTC)

	for _, job := range schedules {
		_, err := scheduler.Cron(job.Expression).Do(cronTask, job)
		if err != nil {
			logger.StderrWithSource.Error(err.Error())
			return
		}
	}

	numJobs := len(scheduler.Jobs())

	if numJobs == 0 {
		logger.StderrWithSource.Warn("no cron jobs where registered")
		os.Exit(1)
	}

	logger.Stdout.Info("starting scheduler", slog.Int("num_jobs", numJobs))

	scheduler.StartBlocking()
}
