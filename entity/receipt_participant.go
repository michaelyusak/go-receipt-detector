package entity

import (
	hEntity "github.com/michaelyusak/go-helper/entity"
)

type ReceiptParticipant struct {
	ParticipantId             int64            `json:"participant_id"`
	ParticipantName           string           `json:"participant_name"`
	ReceiptId                 int64            `json:"receipt_id"`
	NoticeInterval            hEntity.Duration `json:"notice_interval"`
	LastNotice                int64            `json:"last_notice"`
	CreatedAt                 int64            `json:"created_at"`
	UpdatedAt                 *int64           `json:"updated_at,omitempty"`
	DeletedAt                 *int64           `json:"deleted_at,omitempty"`
}
