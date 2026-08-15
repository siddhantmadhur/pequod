package content

type Client interface {
	Fetch(FetchParams, any) error
	SearchMovies(SearchParam) (MovieSearchResponse, error)
	SearchShows(SearchParam) (ShowSearchResponse, error)
	GetSeasonInformation(int, int) (SeriesDetails, error)
	GetEpisodeInformation(int, int, int) (SeriesDetails, error)
	Authenticate() bool
}

func NewClient(client Client) (Client, error) {
	return client, nil
}
