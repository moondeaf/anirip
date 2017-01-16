package crunchyroll

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sdwolfe32/anirip/anirip"
)

// Show contins show metadata and child seasons
type Show struct {
	Title   string
	AdID    string
	Path    string
	URL     string
	Seasons []Season
}

// Scrape appends all the seasons/episodes found for the show
func (s *Show) Scrape(showURL string, cookies []*http.Cookie) error {
	// Gets the HTML of the show page
	showResponse, err := anirip.GetHTTPResponse("GET", showURL, nil, nil, cookies)
	if err != nil {
		return err
	}

	// Creates a goquery document for scraping
	showDoc, err := goquery.NewDocumentFromResponse(showResponse)
	if err != nil {
		return anirip.Error{Message: "There was an error while accessing the show page", Err: err}
	}

	// Sets Title, and Path and URL on our show object
	s.Title = showDoc.Find("#container > h1 > span").First().Text()
	s.URL = showURL
	s.Path = strings.Replace(s.URL, "http://www.crunchyroll.com", "", 1) // Removes the host so we have just the path

	// Searches first for the search div
	showDoc.Find("ul.list-of-seasons.cf").Each(func(i int, seasonList *goquery.Selection) {
		seasonList.Find("li.season").Each(func(i2 int, episodeList *goquery.Selection) {
			// Adds a new season to the show containing all information
			seasonTitle, _ := episodeList.Find("a").First().Attr("title")

			// Adds the title minus any "Episode XX" for shows that only have one season
			s.Seasons = append(s.Seasons, Season{
				Title: strings.SplitN(seasonTitle, " Episode ", 2)[0],
			})

			// Within that season finds all episodes
			episodeList.Find("div.wrapper.container-shadow.hover-classes").Each(func(i3 int, episode *goquery.Selection) {
				// Appends all new episode information to newly appended season
				episodeTitle := strings.TrimSpace(strings.Replace(episode.Find("span.series-title.block.ellipsis").First().Text(), "\n", "", 1))
				episodeNumber, _ := strconv.ParseFloat(strings.Replace(episodeTitle, "Episode ", "", 1), 64)
				episodePath, _ := episode.Find("a").First().Attr("href")
				episodeID, _ := strconv.Atoi(episodePath[len(episodePath)-6:])
				s.Seasons[i2].Episodes = append(show.Seasons[i2].Episodes, Episode{
					ID:     episodeID,
					Title:  episodeTitle,
					Number: episodeNumber,
					Path:   episodePath,
					URL:    "http://www.crunchyroll.com" + episodePath,
				})
			})
		})
	})

	// Re-arranges seasons and episodes in the shows object so we have first to last
	tempSeasonArray := []Season{}
	for i := len(s.Seasons) - 1; i >= 0; i-- {
		// First sort episodes from first to last
		tempEpisodesArray := []Episode{}
		for n := len(s.Seasons[i].Episodes) - 1; n >= 0; n-- {
			tempEpisodesArray = append(tempEpisodesArray, s.Seasons[i].Episodes[n])
		}
		// Lets not bother appending anything if there are no episodes in the season
		if len(tempEpisodesArray) > 0 {
			tempSeasonArray = append(tempSeasonArray, Season{
				Title:    s.Seasons[i].Title,
				Length:   len(tempEpisodesArray),
				Episodes: tempEpisodesArray,
			})
		}
	}
	s.Seasons = tempSeasonArray

	// Assigns each season a number and episode a filename
	for k1, season := range s.Seasons {
		s.Seasons[k1].Number = k1 + 1
		for k2, episode := range season.Episodes {
			s.Seasons[k2].Episodes[k1].FileName = anirip.GenerateEpisodeFileName(s.Title, s.Seasons[k2].Number, episode.Number, "")
		}
	}

	// TODO Filter out episodes that aren't yet released (ex One Piece)
	return nil
}

// Gets the title of the show for referencing outside of this lib
func (show *Show) GetTitle() string {
	return show.Title
}

// Re-stores seasons belonging to the show and returns them for iteration
func (show *Show) GetSeasons() anirip.Seasons {
	seasons := []anirip.Season{}
	for i := 0; i < len(show.Seasons); i++ {
		seasons = append(seasons, &show.Seasons[i])
	}
	return seasons
}
