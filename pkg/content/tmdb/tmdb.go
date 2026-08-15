package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/siddhantmadhur/pequod/pkg/content"
)

type Client struct {
	ApiKey string `json:"api_read_access_token"`
}

func (t Client) Fetch(params content.FetchParams, result any) error {
	if params.Method == "" {
		params.Method = "GET"
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	urlStr := fmt.Sprintf("https://api.themoviedb.org/3%s", params.Endpoint)
	if len(params.Queries) > 0 {
		urlStr += "?" + strings.Join(params.Queries, "&")
	}

	req, err := http.NewRequest(params.Method, urlStr, nil)
	if err != nil {
		return err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.ApiKey))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TMDB API returned status code %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (t Client) GetFromId(Id int) (content.Movie, error) {
	var result content.Movie

	err := t.Fetch(content.FetchParams{
		Endpoint: fmt.Sprintf("/movie/%d", Id),
	}, &result)

	return result, err
}
