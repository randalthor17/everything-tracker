package anilist

import (
	"context"
	"everythingtracker/base"

	"github.com/rl404/verniy"
)

type Anime struct {
	base.BaseMedia
}

func FetchAniListAnime(username string) ([]Anime, error) {
	v := verniy.New()
	ctx := context.Background()

	collection, err := v.GetUserAnimeListWithContext(ctx, username)
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
				BaseMedia: base.BaseMedia{
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
