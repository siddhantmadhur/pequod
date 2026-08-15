package tmdb

import (
	"fmt"
	"net/url"

	"github.com/siddhantmadhur/pequod/pkg/content"
)

func (t Client) SearchMovies(param content.SearchParam) (content.MovieSearchResponse, error) {
	queries := []string{"query=" + url.QueryEscape(param.Query)}
	if param.Year > 0 {
		queries = append(queries, fmt.Sprintf("primary_release_year=%d", param.Year))
	}

	var response content.MovieSearchResponse
	err := t.Fetch(content.FetchParams{
		Endpoint: "/search/movie",
		Queries:  queries,
	}, &response)
	if err != nil {
		return content.MovieSearchResponse{}, err
	}

	return response, nil
}

func (t Client) SearchShows(param content.SearchParam) (content.ShowSearchResponse, error) {
	queries := []string{"query=" + url.QueryEscape(param.Query)}
	if param.Year > 0 {
		queries = append(queries, fmt.Sprintf("first_air_date_year=%d", param.Year))
	}

	var response content.ShowSearchResponse
	err := t.Fetch(content.FetchParams{
		Endpoint: "/search/tv",
		Queries:  queries,
	}, &response)
	if err != nil {
		return content.ShowSearchResponse{}, err
	}

	return response, nil
}
