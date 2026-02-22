package anilist

import (
	"everythingtracker/base"

	"github.com/rl404/verniy"
)

type Manga struct {
	base.BaseMedia
}

func FetchAniListManga(username string) ([]Manga, error) {
	v := verniy.New()

	collection, err := v.GetUserMangaList(username)
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
				base.BaseMedia{
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

func SearchAnilistManga(query string, searchCount int) ([]Manga, error) {
	v := verniy.New()

	searchPage, err := v.SearchManga(verniy.PageParamMedia{Search: query}, 1, searchCount)
	if err != nil {
		return nil, err
	}

	var res []Manga
	for _, media := range searchPage.Media {
		res = append(res, Manga{
			base.BaseMedia{
				Title: ExtractTitle(media.ID, &media),
			},
		})
	}

	return res, nil
}
