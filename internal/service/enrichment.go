package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type EnrichmentClient struct {
	client *http.Client
	log    *logrus.Logger
}

func NewEnrichmentClient(log *logrus.Logger) *EnrichmentClient {
	return &EnrichmentClient{
		client: &http.Client{Timeout: 5 * time.Second},
		log:    log,
	}
}

type EnrichmentResult struct {
	Age         int
	Gender      string
	Nationality string
	Errors      []error
}

func (c *EnrichmentClient) GetAge(ctx context.Context, name string) (int, error) {
	resp, err := c.client.Get(fmt.Sprintf("https://api.agify.io/?name=%s", name))
	if err != nil {
		return 0, fmt.Errorf("agify request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("agify returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read agify response: %w", err)
	}

	var data struct{ Age int }
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("failed to parse agify response: %w", err)
	}
	return data.Age, nil
}

func (c *EnrichmentClient) GetGender(ctx context.Context, name string) (string, error) {
	resp, err := c.client.Get(fmt.Sprintf("https://api.genderize.io/?name=%s", name))
	if err != nil {
		return "", fmt.Errorf("genderize request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("genderize returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read genderize response: %w", err)
	}

	var data struct{ Gender string }
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to parse genderize response: %w", err)
	}
	return data.Gender, nil
}

func (c *EnrichmentClient) GetNationality(ctx context.Context, name string) (string, error) {
	resp, err := c.client.Get(fmt.Sprintf("https://api.nationalize.io/?name=%s", name))
	if err != nil {
		return "", fmt.Errorf("nationalize request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nationalize returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read nationalize response: %w", err)
	}

	var data struct {
		Country []struct {
			CountryID string `json:"country_id"`
		} `json:"country"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to parse nationalize response: %w", err)
	}
	if len(data.Country) == 0 {
		return "", nil
	}
	return data.Country[0].CountryID, nil
}
