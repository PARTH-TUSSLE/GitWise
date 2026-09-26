package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// GraphQLCalendarResponse represents the decoded contribution calendar from GitHub GraphQL API.
type GraphQLCalendarResponse struct {
	TotalContributions int
	Weeks              []GraphQLContributionWeek
}

type GraphQLContributionWeek struct {
	ContributionDays []GraphQLContributionDay
}

type GraphQLContributionDay struct {
	Date              string // "YYYY-MM-DD"
	ContributionCount int
	ContributionLevel string // "NONE", "FIRST_QUARTILE", "SECOND_QUARTILE", "THIRD_QUARTILE", "FOURTH_QUARTILE"
}

// rawGraphQLPayload matches GitHub's nested GraphQL JSON response.
type rawGraphQLPayload struct {
	Data struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []struct {
							Date              string `json:"date"`
							ContributionCount int    `json:"contributionCount"`
							ContributionLevel string `json:"contributionLevel"`
						} `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

const calendarQuery = `
query($username: String!) {
  user(login: $username) {
    contributionsCollection {
      contributionCalendar {
        totalContributions
        weeks {
          contributionDays {
            date
            contributionCount
            contributionLevel
          }
        }
      }
    }
  }
}
`

// GetContributionCalendar queries GitHub's GraphQL API for the user's 52-week activity calendar.
// If unauthenticated or token is missing, it returns ErrGraphQLUnavailable for graceful fallback.
func (c *Client) GetContributionCalendar(ctx context.Context, username string) (*GraphQLCalendarResponse, error) {
	if c.token == "" {
		return nil, ErrGraphQLUnavailable
	}

	reqBody := map[string]interface{}{
		"query": calendarQuery,
		"variables": map[string]string{
			"username": username,
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal graphql query: %w", err)
	}

	resp, err := c.Do(ctx, "POST", c.graphqlURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		c.logger.Debug("GraphQL unauthenticated or forbidden; falling back to event calendar",
			slog.Int("status", resp.StatusCode),
		)
		return nil, ErrGraphQLUnavailable
	}

	if err := c.checkResponseStatus(resp, "graphql_calendar", username); err != nil {
		return nil, err
	}

	var payload rawGraphQLPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode graphql response: %w", err)
	}

	if len(payload.Errors) > 0 {
		c.logger.Debug("GraphQL returned errors", slog.String("error", payload.Errors[0].Message))
		return nil, fmt.Errorf("graphql error: %s", payload.Errors[0].Message)
	}

	rawCalendar := payload.Data.User.ContributionsCollection.ContributionCalendar
	weeks := make([]GraphQLContributionWeek, len(rawCalendar.Weeks))
	for wIdx, w := range rawCalendar.Weeks {
		days := make([]GraphQLContributionDay, len(w.ContributionDays))
		for dIdx, d := range w.ContributionDays {
			days[dIdx] = GraphQLContributionDay{
				Date:              d.Date,
				ContributionCount: d.ContributionCount,
				ContributionLevel: d.ContributionLevel,
			}
		}
		weeks[wIdx] = GraphQLContributionWeek{
			ContributionDays: days,
		}
	}

	return &GraphQLCalendarResponse{
		TotalContributions: rawCalendar.TotalContributions,
		Weeks:              weeks,
	}, nil
}
