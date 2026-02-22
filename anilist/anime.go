package anilist

import (
	"everythingtracker/base"

	"github.com/rl404/verniy"
)

type Anime struct {
	base.BaseMedia
}

func FetchAniListAnime(username string) ([]Anime, error) {
	v := verniy.New()

	collection, err := v.GetUserAnimeList(username)
	if err != nil {
		return nil, err
	}

	var items []Anime

	for _, list := range collection {
		for _, entry := range list.Entries {
			// for not yet released anime, ProgressTotal is not available, so we set it to 0
			progressTotal := 0.0

			if entry.Media != nil && entry.Media.Episodes != nil {
				progressTotal = float64(*entry.Media.Episodes)
			} else {
				print("No episode count found for media id ", entry.ID, "\n")
				print("Using 0 as fallback for progress_total\n")
			}

			item := Anime{
				base.BaseMedia{
					Title:           ExtractTitle(entry.ID, entry.Media),
					ExternalID:      entry.ID,
					Status:          MapAniListStatus(string(*entry.Status), true),
					ProgressCurrent: float64(*entry.Progress),
					ProgressTotal:   progressTotal,
					ProgressUnit:    "ep",
				},
			}
			items = append(items, item)
		}
	}
	return items, nil
}

func SearchAnilistAnime(query string, searchCount int) ([]Anime, error) {
	v := verniy.New()

	searchPage, err := v.SearchAnime(verniy.PageParamMedia{Search: query}, 1, searchCount)
	if err != nil {
		return nil, err
	}

	var res []Anime
	for _, media := range searchPage.Media {
		res = append(res, Anime{
			base.BaseMedia{
				Title:      ExtractTitle(media.ID, &media),
				ExternalID: media.ID,
			},
		})
	}

	return res, nil
}
