package library

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

func getShowInformation(fullPath string, name string) (string, int, int, error) {
	tokens := strings.Split(fullPath, "/")
	if len(tokens) < 3 {
		return "", 0, 0, errors.New("Not enough information")
	}
	getNumber, err := regexp.Compile("[0-9]+")
	if err != nil {
		return "", 0, 0, err
	}

	getSeasonString, err := regexp.Compile("s[0-9]+|S[0-9]+|Season [0-9]+")
	getEpisodeString, err := regexp.Compile("e[0-9]+|E[0-9]+|Episode [0-9]+")

	if err != nil {
		return "", 0, 0, err
	}

	seasonString := getSeasonString.FindString(fullPath)
	episodeString := getEpisodeString.FindString(name)

	season, err := strconv.Atoi(getNumber.FindString(seasonString))
	episode, err := strconv.Atoi(getNumber.FindString(episodeString))

	return strings.ReplaceAll(tokens[1], ".", " "), season, episode, err

}

// Gets the depth of a file relative to the "root". i.e 1 for shows, 2 for seasons and 3 for episodes
func getFileDepth(rootPath string, completePath string) int {
	relativePath := strings.Replace(completePath, rootPath, "", 1)
	rawPaths := strings.Split(relativePath, "/")
	paths := []string{}
	for _, path := range rawPaths {
		if path != "" {
			paths = append(paths, path)
		}
	}
	return len(paths)
}
