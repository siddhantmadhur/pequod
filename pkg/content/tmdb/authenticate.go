package tmdb

import "github.com/siddhantmadhur/pequod/pkg/content"

func (t Client) Authenticate() bool {
	var response struct {
		Success bool `json:"success"`
	}
	err := t.Fetch(content.FetchParams{
		Endpoint: "/authentication",
	}, &response)

	if err != nil {
		return false
	}

	return response.Success
}
