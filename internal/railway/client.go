package railway

import (
	"fmt"
	"main/internal/schedule"
	"net/http"

	_ "github.com/Khan/genqlient/generate"
	"github.com/Khan/genqlient/graphql"
	"golang.org/x/exp/slices"
)

type authedTransport struct {
	token   string
	wrapped http.RoundTripper
}

type railwayClient struct {
	graphql.Client
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	return t.wrapped.RoundTrip(req)
}

func NewAuthedClient(token string) *railwayClient {
	httpClient := http.Client{
		Transport: &authedTransport{
			token:   token,
			wrapped: http.DefaultTransport,
		},
	}

	return &railwayClient{
		graphql.NewClient("https://backboard.railway.app/graphql/v2", &httpClient),
	}
}

func (rc *railwayClient) GetFriendlyName(id string) (string, error) {
	resp, err := Service(rc, id)
	if err != nil {
		return "undefined", err
	}

	return resp.Service.Name, nil
}

func (rc *railwayClient) GetLatestDeploymentID(schedule schedule.Schedule) (string, error) {
	resp, err := Deployments(rc, 50, &DeploymentListInput{
		ProjectId:     schedule.ProjectID,
		EnvironmentId: schedule.EnvironmentID,
		ServiceId:     schedule.ServiceID,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Deployments.Edges) == 0 {
		return "", fmt.Errorf("no deployments for service found")
	}

	acceptedStatuses := []string{"SUCCESS", "COMPLETED"}

	for _, deployment := range resp.Deployments.Edges {
		if slices.Contains(acceptedStatuses, string(deployment.Node.Status)) == true {
			return deployment.Node.Id, nil
		}
	}

	return "", fmt.Errorf("no deployments found with status SUCCESS or COMPLETED for service id: %v", schedule.ServiceID)
}
