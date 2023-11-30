package aggregates

import (
	"tryffel.net/go/virtualpaper/services/process"
	"tryffel.net/go/virtualpaper/services/search"
)

// swagger:response SystemInfo
type SystemInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	GoVersion string `json:"go_version"`

	ImagemagickVersion string `json:"imagemagick_version"`
	TesseractVersion   string `json:"tesseract_version"`
	PopplerInstalled   bool   `json:"poppler_installed"`
	PandocInstalled    bool   `json:"pandoc_installed"`

	NumCpu     int    `json:"number_cpus"`
	ServerLoad string `json:"server_load"`
	Uptime     string `json:"uptime"`

	DocumentsInQueue            int    `json:"documents_queued"`
	DocumentsProcessedToday     int    `json:"documents_processed_today"`
	DocumentsProcessedLastWeek  int    `json:"documents_processed_past_week"`
	DocumentsProcessedLastMonth int    `json:"documents_processed_past_month"`
	DocumentsTotal              int    `json:"documents_total"`
	DocumentsTotalSize          int64  `json:"documents_total_size"`
	DocumentsTotalSizeString    string `json:"documents_total_size_string"`

	ProcessingStatus   []process.QueueStatus `json:"processing_queue"`
	SearchEngineStatus search.EngineStatus   `json:"search_engine_status"`

	ProcessingEnabled bool `json:"processing_enabled"`
	CronJobsEnabled   bool `json:"cronjobs_enabled"`
}
