// Package base holds base media declarations
package base

import "gorm.io/gorm"

type MediaStatus string

const (
	StatusPlanningWatch MediaStatus = "Plan to Watch"
	StatusWatching      MediaStatus = "Watching"
	StatusPlanningRead  MediaStatus = "Plan to Read"
	StatusReading       MediaStatus = "Reading"
	StatusCompleted     MediaStatus = "Completed"
	StatusDropped       MediaStatus = "Dropped"
	StatusPaused        MediaStatus = "Paused"
)

type BaseMedia struct {
	gorm.Model
	Title           string      `json:"title"`
	ExternalID      int         `json:"external_id" gorm:"uniqueIndex"`
	Status          MediaStatus `json:"status"`
	ProgressCurrent float64     `json:"progress_current"`
	ProgressTotal   float64     `json:"progress_total"`
	ProgressUnit    string      `json:"progress_unit"` // ep, ch, percent, min
}
