package entity

type ParticipantContact struct {
	ContactId     int64  `json:"contact_id"`
	ParticipantId int64  `json:"participant_id"`
	ContactType   string `json:"contact_type" binding:"required"`
	ContactValue  string `json:"contact_value" binding:"required"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     *int64 `json:"updated_at,omitempty"`
	DeletedAt     *int64 `json:"deleted_at,omitempty"`
}

type GetAllowedContactTypesResponse struct {
	AllowedContactTypes []string `json:"allowed_contact_types"`
}
