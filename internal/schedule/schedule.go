package schedule

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"github.com/robfig/cron/v3"
)

type action string

const (
	ActionRestart  action = "restart"
	ActionRedeploy action = "redeploy"
)

type Schedule struct {
	ServiceID     string `schema:"serviceID,required"`
	ProjectID     string `schema:"projectID,required"`
	EnvironmentID string `schema:"environmentID,required"`

	Action     action `schema:"action,required"`
	Expression string `schema:"expression,required"`
}

func ParseFromEnv(EnvPrefix string) ([]Schedule, error) {
	var schedules []Schedule

	for _, env := range os.Environ() {
		env = strings.TrimSpace(env)

		if strings.HasPrefix(env, EnvPrefix) == false {
			continue
		}

		kv := strings.SplitN(env, "=", 2)

		if len(kv) != 2 {
			return nil, fmt.Errorf("environment key value pair: %q invalid; expected k=v format", env)
		}

		value := strings.TrimSpace(kv[1])

		if valueUnquoted, err := strconv.Unquote(value); err == nil {
			value = valueUnquoted
		}

		// extra check to make sure the schedule string only contains k=v pairs
		for _, param := range strings.Split(value, "&") {
			if len(strings.SplitN(param, "=", 2)) != 2 {
				return nil, fmt.Errorf("schedule key value pair: %q invalid; expected k=V format", param)
			}
		}

		values, err := url.ParseQuery(value)
		if err != nil {
			return nil, err
		}

		var schedule Schedule

		if err := schema.NewDecoder().Decode(&schedule, values); err != nil {
			return nil, err
		}

		if schedule.Action != ActionRestart && schedule.Action != ActionRedeploy {
			return nil, fmt.Errorf("invalid action: %q", schedule.Action)
		}

		schedule.Expression = strings.ToLower(schedule.Expression)

		cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

		if _, err := cronParser.Parse(schedule.Expression); err != nil {
			return nil, err
		}

		if _, err := uuid.Parse(schedule.ProjectID); err != nil {
			return nil, fmt.Errorf("projectID is not a valid UUID: %q", schedule.ProjectID)
		}

		if _, err := uuid.Parse(schedule.EnvironmentID); err != nil {
			return nil, fmt.Errorf("environmentID is not a valid UUID: %q", schedule.EnvironmentID)
		}

		if _, err := uuid.Parse(schedule.ServiceID); err != nil {
			return nil, fmt.Errorf("serviceID is not a valid UUID: %q", schedule.ServiceID)
		}

		schedules = append(schedules, schedule)
	}

	return schedules, nil
}
