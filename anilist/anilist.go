// Package anilist provides functions to extract and map data from AniList media entries
package anilist

import (
	"everythingtracker/base"

	"github.com/rl404/verniy"
)

// ExtractTitle extracts the title from AniList media entry
// Prefers English title, falls back to Romaji, then "Unknown Title"
func ExtractTitle(mediaID int, media *verniy.Media) string {
	title := "Unknown Title"

	if media != nil && media.Title != nil {
		if media.Title.English != nil {
			title = *media.Title.English
		} else if media.Title.Romaji != nil {
			title = *media.Title.Romaji
		} else {
			print("No title found for media id ", mediaID, "\n", "Using 'Unknown Title' as fallback\n")
		}
	}

	return title
}

// MapAniListStatus maps AniList status to internal MediaStatus
func MapAniListStatus(status string, isAnime bool) base.MediaStatus {
	switch status {
	case "CURRENT":
		if isAnime {
			return base.StatusWatching
		}
		return base.StatusReading
	case "PLANNING":
		if isAnime {
			return base.StatusPlanningWatch
		}
		return base.StatusPlanningRead
	case "COMPLETED":
		return base.StatusCompleted
	case "DROPPED":
		return base.StatusDropped
	case "PAUSED":
		return base.StatusPaused
	default:
		return base.StatusPlanningWatch
	}
}
