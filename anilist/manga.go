package anilist

import (
	"context"
	"everythingtracker/base"

	"github.com/rl404/verniy"
)

type Manga struct {
	base.BaseMedia
}

func FetchAniListManga(username string) ([]Manga, error) {
	v := verniy.New()
	ctx := context.Background()

	collection, err := v.GetUserMangaListWithContext(ctx, username)
	if err != nil {
		return nil, err
	}

	var items []Manga

	for _, list := range collection {
		for _, entry := range list.Entries {
			// for ongoing manga, Anilist doesn't track total chapters released, so we set it to chapters read
			progressTotal := 0.0

			if entry.Media != nil && entry.Media.Chapters != nil {
				progressTotal = float64(*entry.Media.Chapters)
			} else {
				print("No chapter count found for media id ", entry.ID, "\n")
				print("Using chapters read as fallback for progress_total\n")
				progressTotal = float64(*entry.Progress)
			}

			item := Manga{
				BaseMedia: base.BaseMedia{
					Title:           ExtractTitle(entry.ID, entry.Media),
					ExternalID:      entry.ID,
					Status:          MapAniListStatus(string(*entry.Status), false),
					ProgressCurrent: float64(*entry.Progress),
					ProgressTotal:   progressTotal,
					ProgressUnit:    "ch",
				},
			}
			items = append(items, item)
		}
	}
	return items, nil
}
