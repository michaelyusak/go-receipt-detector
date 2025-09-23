package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"receipt-detector/entity"
	"receipt-detector/helper"
	"receipt-detector/repository"
	"strconv"
)

type participantContacts struct {
	dbtx repository.DBTX
}

func NewParticipantContacts(dbtx repository.DBTX) *participantContacts {
	return &participantContacts{
		dbtx: dbtx,
	}
}

func (r *participantContacts) NewTx(tx *sql.Tx) repository.ParticipantContacts {
	return &participantContacts{
		dbtx: tx,
	}
}

func (r *participantContacts) InsertMany(ctx context.Context, contacts []entity.ParticipantContact) error {
	q := `
		INSERT
		INTO participant_contacts (participant_id, contact_type, contact_value, created_at)
		VALUES 
	`
	now := helper.NowUnixMilli()
	args := []any{}

	for i, contact := range contacts {
		offset := i * 4

		participantIdIdx := strconv.Itoa(offset + 1)
		contactTypeIdx := strconv.Itoa(offset + 2)
		contactValueIdx := strconv.Itoa(offset + 3)
		createdAtIdx := strconv.Itoa(offset + 4)

		q := `($` + participantIdIdx +
			`, $` + contactTypeIdx +
			`, $` + contactValueIdx +
			`, $` + createdAtIdx + `)`

		if i < len(contacts)-1 {
			q += `, `
		}

		args = append(args, contact.ParticipantId)
		args = append(args, contact.ContactType)
		args = append(args, contact.ContactValue)
		args = append(args, now)
	}

	_, err := r.dbtx.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("repository][postgres][participantContacts][InsertMany][dbtx.ExecContext] %w", err)
	}

	return nil
}

func (r *participantContacts) GetByParticipantId(ctx context.Context, participantId int64) ([]entity.ParticipantContact, error) {
	q := `
		SELECT contact_id, participant_id, contact_type, contact_value, created_at, updated_at
		FROM participant_contacts
		WHERE participant_id = $1
			AND deleted_at IS NULL
	`

	contacts := []entity.ParticipantContact{}

	rows, err := r.dbtx.QueryContext(ctx, q, participantId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return contacts, nil
		}

		return nil, fmt.Errorf("repository][postgres][participantContacts][GetByParticipantId][dbtx.QueryContext] %w", err)
	}

	for rows.Next() {
		var contact entity.ParticipantContact

		err := rows.Scan(
			&contact.ContactId,
			&contact.ParticipantId,
			&contact.ContactType,
			&contact.ContactValue,
			&contact.CreatedAt,
			&contact.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("repository][postgres][participantContacts][GetByParticipantId][rows.Scan] %w", err)
		}

		contacts = append(contacts, contact)
	}

	return contacts, nil
}
