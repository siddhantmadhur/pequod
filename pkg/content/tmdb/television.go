package tmdb

import (
	"fmt"

	"github.com/siddhantmadhur/pequod/pkg/content"
)

func (t Client) GetSeasonInformation(seriesId int, seasonNo int) (content.SeriesDetails, error) {

	var response content.SeriesDetails

	err := t.Fetch(content.FetchParams{
		Endpoint: fmt.Sprintf("/tv/%d/season/%d", seriesId, seasonNo),
	}, &response)

	if err != nil {
		return content.SeriesDetails{}, err
	}

	return response, nil
}

func (t Client) GetEpisodeInformation(seriesId int, seasonNo int, episodeNo int) (content.SeriesDetails, error) {

	var response content.SeriesDetails

	err := t.Fetch(content.FetchParams{
		Endpoint: fmt.Sprintf("/tv/%d/season/%d/episode/%d", seriesId, seasonNo, episodeNo),
	}, &response)

	if err != nil {
		return content.SeriesDetails{}, err
	}

	return response, nil
}
