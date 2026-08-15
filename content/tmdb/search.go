package tmdb

import (
	"fmt"
	"strings"

	"github.com/siddhantmadhur/pequod/content"
)

func (t Client) SearchMovies(param content.SearchParam) (content.MovieSearchResponse, error) {

	var response content.MovieSearchResponse
	err := t.Fetch(content.FetchParams{
		Endpoint: "/search/movie",
		Queries:  []string{"query=" + strings.ReplaceAll(param.Query, " ", "%20"), fmt.Sprintf("first_air_date_year=%d", param.Year)},
	}, &response)
	if err != nil {
		return content.MovieSearchResponse{}, err
	}

	return response, nil
}

func (t Client) SearchShows(param content.SearchParam) (content.ShowSearchResponse, error) {

	var response content.ShowSearchResponse
	err := t.Fetch(content.FetchParams{
		Endpoint: "/search/tv",
		Queries:  []string{"query=" + strings.ReplaceAll(param.Query, " ", "%20"), fmt.Sprintf("first_air_date_year=%d", param.Year)},
	}, &response)

	if err != nil {
		return content.ShowSearchResponse{}, err
	}

	return response, nil
}
