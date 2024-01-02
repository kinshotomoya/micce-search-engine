package model

type UpsertCondition struct {
	SpotId         string
	UpdatedAt      string
	VespaUpdatedAt *string
	IsVespaUpdated bool
	IndexStatus    INDEX_STATUS
}

type INDEX_STATUS string

const (
	READY      INDEX_STATUS = "READY"
	PROCESSING INDEX_STATUS = "PROCESSING"
	COMPLETED  INDEX_STATUS = "COMPLETED"
)
