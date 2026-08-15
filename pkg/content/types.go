package content

type Genre struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type FetchParams struct {
	Method   string   `json:"method"`
	Endpoint string   `json:"endpoint"`
	Queries  []string `json:"queries"`
}

type Movie struct {
	Id               int      `json:"id"`
	Title            string   `json:"title"`
	Genres           []Genre  `json:"genres"`
	ImdbId           string   `json:"imdb_id"`
	OriginCountry    []string `json:"origin_country"`
	Overview         string   `json:"overview"`
	OriginalLanguage string   `json:"original_language"`
	OriginalTitle    string   `json:"original_title"`
	ReleaseDate      string   `json:"release_date"`
	Status           string   `json:"status"`
	Tagline          string   `json:"tagline"`
	PosterPath       string   `json:"poster_path"`
}

type Show struct {
	Id            int      `json:"id"`
	Name          string   `json:"name"`
	Genres        []Genre  `json:"genres"`
	OriginCountry []string `json:"origin_country"`
	PosterPath    string   `json:"poster_path"`
	Overview      string   `json:"overview"`
	SeasonNumber  int      `json:"season_number"`
	EpisodeNumber int      `json:"episode_number"`
}

type MovieSearchResponse struct {
	Page    int     `json:"page"`
	Results []Movie `json:"results"`
}

type ShowSearchResponse struct {
	Page    int    `json:"page"`
	Results []Show `json:"results"`
}

type SearchParam struct {
	Query string `json:"query"`
	Year  int32  `json:"year"`
}

type SeriesDetails struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	SeasonNumber int    `json:"season_number"`
}
