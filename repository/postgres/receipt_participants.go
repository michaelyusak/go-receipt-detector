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

type receiptParticipants struct {
	dbtx repository.DBTX
}

func NewReceiptParticipants(dbtx repository.DBTX) *receiptParticipants {
	return &receiptParticipants{
		dbtx: dbtx,
	}
}

func (r *receiptParticipants) NewTx(tx *sql.Tx) repository.ReceiptParticipants {
	return &receiptParticipants{
		dbtx: tx,
	}
}

func (r *receiptParticipants) InsertMany(ctx context.Context, participants []entity.ReceiptParticipant) error {
	q := `
		INSERT
		INTO receipt_participants (participant_name, receipt_id, notice_interval, created_at)
		VALUES 
	`

	now := helper.NowUnixMilli()
	args := []any{}

	for i, participant := range participants {
		offset := i * 4

		participantNameIdx := strconv.Itoa(offset + 1)
		receiptIdIdx := strconv.Itoa(offset + 2)
		noticeIntervalIdx := strconv.Itoa(offset + 3)
		createdAtIdx := strconv.Itoa(offset + 4)

		q += `($` + participantNameIdx +
			`, $` + receiptIdIdx +
			`, $` + noticeIntervalIdx +
			`, $` + createdAtIdx + `)`

		if i < len(participants)-1 {
			q += `, `
		}

		args = append(args, participant.ParticipantName)
		args = append(args, participant.ReceiptId)
		args = append(args, participant.NoticeInterval)
		args = append(args, now)
	}

	_, err := r.dbtx.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("repository][postgres][receiptParticipants][InsertMany][dbtx.ExecContext] %w", err)
	}

	return nil
}

func (r *receiptParticipants) GetByReceiptId(ctx context.Context, receiptId int64) ([]entity.ReceiptParticipant, error) {
	q := `
		SELECT participant_id, participant_name, receipt_id, notice_interval, last_notice, created_at, updated_at
		FROM receipt_participants
		WHERE receipt_id = $1
			AND deleted_at IS NULL
	`

	participants := []entity.ReceiptParticipant{}

	rows, err := r.dbtx.QueryContext(ctx, q, receiptId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return participants, nil
		}

		return nil, fmt.Errorf("repository][postgres][receiptParticipants][GetByReceiptId][dbtx.QueryContext] %w", err)
	}

	for rows.Next() {
		var participant entity.ReceiptParticipant

		err := rows.Scan(
			&participant.ParticipantId,
			&participant.ParticipantName,
			&participant.ReceiptId,
			&participant.NoticeInterval,
			&participant.LastNotice,
			&participant.CreatedAt,
			&participant.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("repository][postgres][receiptParticipants][GetByReceiptId][rows.Scan] %w", err)
		}

		participants = append(participants, participant)
	}

	return participants, nil
}
