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
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	hEntity "github.com/michaelyusak/go-helper/entity"
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

func (r *receiptParticipants) InsertMany(ctx context.Context, receiptId int64, participants []entity.ReceiptParticipant) ([]int64, error) {
	q := `
		INSERT
		INTO receipt_participants (participant_name, receipt_id, notifying, notice_interval, created_at)
		VALUES 
	`

	now := helper.NowUnixMilli()
	args := []any{}

	for i, participant := range participants {
		offset := i * 5

		participantNameIdx := strconv.Itoa(offset + 1)
		receiptIdIdx := strconv.Itoa(offset + 2)
		notifyingIdx := strconv.Itoa(offset + 3)
		noticeIntervalIdx := strconv.Itoa(offset + 4)
		createdAtIdx := strconv.Itoa(offset + 5)

		q += `($` + participantNameIdx +
			`, $` + receiptIdIdx +
			`, $` + notifyingIdx +
			`, $` + noticeIntervalIdx +
			`, $` + createdAtIdx + `)`

		if i < len(participants)-1 {
			q += `, `
		}

		args = append(args, participant.ParticipantName)
		args = append(args, receiptId)
		args = append(args, participant.Notifying)
		args = append(args, time.Duration(participant.NoticeInterval).Milliseconds())
		args = append(args, now)
	}

	q += `RETURNING participant_id`

	participantIds := []int64{}

	rows, err := r.dbtx.QueryContext(ctx, q, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil, fmt.Errorf("[repository][postgres][receiptParticipants][InsertMany][dbtx.QueryContext] %w: %v", repository.ErrUniqueViolation, err)
			}
		}

		return participantIds, fmt.Errorf("[repository][postgres][receiptParticipants][InsertMany][dbtx.QueryContext] %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var participantId int64

		err := rows.Scan(&participantId)
		if err != nil {
			return nil, fmt.Errorf("[repository][postgres][receiptParticipants][InsertMany][rows.Scan] %w", err)
		}

		participantIds = append(participantIds, participantId)
	}

	return participantIds, nil
}

func (r *receiptParticipants) GetByReceiptId(ctx context.Context, receiptId int64) ([]entity.ReceiptParticipant, error) {
	q := `
		SELECT participant_id, participant_name, receipt_id, notifying, notice_interval, last_notice, created_at, updated_at
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

		return nil, fmt.Errorf("[repository][postgres][receiptParticipants][GetByReceiptId][dbtx.QueryContext] %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var participant entity.ReceiptParticipant
		var noticeIntervalInt64 int64

		err = rows.Scan(
			&participant.ParticipantId,
			&participant.ParticipantName,
			&participant.ReceiptId,
			&participant.Notifying,
			&noticeIntervalInt64,
			&participant.LastNotice,
			&participant.CreatedAt,
			&participant.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("[repository][postgres][receiptParticipants][GetByReceiptId][rows.Scan] %w", err)
		}

		participant.NoticeInterval = hEntity.Duration(noticeIntervalInt64 * int64(time.Millisecond))

		participants = append(participants, participant)
	}

	return participants, nil
}

func (r *receiptParticipants) GetByParticipantId(ctx context.Context, participantId int64) (*entity.ReceiptParticipant, error) {
	q := `
		SELECT participant_id, participant_name, receipt_id, notifying, notice_interval, last_notice, created_at, updated_at
		FROM receipt_participants
		WHERE participant_id = $1
			AND deleted_at IS NULL
	`

	var participant entity.ReceiptParticipant

	err := r.dbtx.QueryRowContext(ctx, q, participantId).Scan(
		&participant.ParticipantId,
		&participant.ParticipantName,
		&participant.ReceiptId,
		&participant.Notifying,
		&participant.NoticeInterval,
		&participant.LastNotice,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("[repository][postgres][receiptParticipants][GetByParticipantId][dbtx.QueryRowContext] %w", err)
	}

	return &participant, nil
}

func (r *receiptParticipants) UpdateOne(ctx context.Context, receiptParticipant entity.ReceiptParticipant) error {
	q := `
		UPDATE receipt_participants
		SET 
	`

	args := []any{}

	q += `notifying = $1, `
	args = append(args, receiptParticipant.Notifying)

	i := 2

	if !receiptParticipant.Notifying || receiptParticipant.NoticeInterval > 0 {
		q += `notice_interval = $` + strconv.Itoa(i) + ", "
		args = append(args, time.Duration(receiptParticipant.NoticeInterval).Milliseconds())
		i++
	}

	if receiptParticipant.ParticipantName != "" {
		q += `participant_name = $` + strconv.Itoa(i) + ", "
		args = append(args, receiptParticipant.ParticipantName)
		i++
	}

	q += `updated_at = $` + strconv.Itoa(i) +
		` WHERE participant_id = $` + strconv.Itoa(i+1)
	args = append(args, helper.NowUnixMilli())
	args = append(args, receiptParticipant.ParticipantId)

	_, err := r.dbtx.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("[repository][postgres][receiptParticipants][UpdateOne][dbtx.ExecContext] %w", err)
	}

	return nil
}

func (r *receiptParticipants) DeleteByParticipantIds(ctx context.Context, participantIds []int64) error {
	q := `
		DELETE FROM receipt_participants
		WHERE participant_id
		IN (
	`

	args := []any{}

	for i, participantId := range participantIds {
		q += `$` + strconv.Itoa(i+1)
		args = append(args, participantId)

		if i != len(participantIds)-1 {
			q += `, `
		}
	}

	q += `)`

	_, err := r.dbtx.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("[repository][postgres][receiptParticipants][DeleteByParticipantIds][dbtx.ExecContext] %w", err)
	}

	return nil
}
