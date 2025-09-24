package entity

import (
	hEntity "github.com/michaelyusak/go-helper/entity"
)

type ReceiptParticipant struct {
	ParticipantId   int64                `json:"participant_id"`
	ParticipantName string               `json:"participant_name" binding:"required"`
	ReceiptId       int64                `json:"receipt_id"`
	Notifying       bool                 `json:"notifying"`
	NoticeInterval  hEntity.Duration     `json:"notice_interval"`
	LastNotice      *int64               `json:"last_notice"`
	Contacts        []ParticipantContact `json:"contacts,omitempty" binding:"required,gte=1,dive"`
	CreatedAt       int64                `json:"created_at"`
	UpdatedAt       *int64               `json:"updated_at,omitempty"`
	DeletedAt       *int64               `json:"deleted_at,omitempty"`
}
}
