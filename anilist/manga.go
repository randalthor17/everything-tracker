package anilist

import (
	"everythingtracker/base"

	"github.com/rl404/verniy"
)

type Manga struct {
	base.BaseMedia
}

// TableName sets the table name for manga
func (Manga) TableName() string {
	return "mangas"
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
				print("No chapter count found for media id ", entry.Media.ID, "\n")
				print("Using chapters read as fallback for progress_total\n")
				progressTotal = float64(*entry.Progress)
			}

			item := Manga{}
			item.Title = ExtractTitle(entry.Media.ID, entry.Media)
			item.ExternalID = entry.Media.ID
			item.Status = MapAniListStatus(string(*entry.Status), false)
			item.ProgressCurrent = float64(*entry.Progress)
			item.ProgressTotal = progressTotal
			item.ProgressUnit = "ch"
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
		item := Manga{}
		item.Title = ExtractTitle(media.ID, &media)
		item.ExternalID = media.ID
		res = append(res, item)
	}

	return res, nil
}


func GetMangaByExternalID(externalID int) (*Manga, error) {
	v := verniy.New()

	media, err := v.GetManga(externalID)
	if err != nil {
		return nil, err
	}
	
	item := Manga{}
	item.Title = ExtractTitle(media.ID, media)
	item.ExternalID = media.ID
	item.ProgressUnit = "ch"
	
	// AniList doesn't reliably track total chapters for ongoing manga
	if media.Chapters != nil && *media.Chapters > 0 {
		item.ProgressTotal = float64(*media.Chapters)
	} else {
		item.ProgressTotal = 0 // Unknown/ongoing
	}
	
	return &item, nil
}

