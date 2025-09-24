package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"receipt-detector/entity"
	"receipt-detector/repository"
	"time"

	hAppconstant "github.com/michaelyusak/go-helper/appconstant"
	hApperror "github.com/michaelyusak/go-helper/apperror"
	hEntity "github.com/michaelyusak/go-helper/entity"
	"github.com/sirupsen/logrus"
)

type receiptParticipant struct {
	receiptParticipantsRepo repository.ReceiptParticipants
	participantContactsRepo repository.ParticipantContacts
	receiptsRepo            repository.Receipts

	transaction repository.Transaction

	logTag string
}

type ReceiptParticipantOpt struct {
	ReceiptParticipantsRepo repository.ReceiptParticipants
	ParticipantContactsRepo repository.ParticipantContacts
	ReceiptsRepo            repository.Receipts

	Transaction repository.Transaction
}

func NewReceiptParticipant(opt ReceiptParticipantOpt) *receiptParticipant {
	return &receiptParticipant{
		receiptParticipantsRepo: opt.ReceiptParticipantsRepo,
		participantContactsRepo: opt.ParticipantContactsRepo,
		receiptsRepo:            opt.ReceiptsRepo,

		transaction: opt.Transaction,

		logTag: "[service][receiptParticipant]",
	}
}

func (s *receiptParticipant) AddParticipantsOneByOne(ctx context.Context, receiptId int64, participants []entity.ReceiptParticipant) error {
	logTag := s.logTag + "[AddParticipantsOneByOne]"

	tx, err := s.transaction.Begin()
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[transaction.Begin] Failed to begin transaction: %v [receipt_id: %v]", logTag, err, receiptId),
		})
	}

	defer func() {
		if p := recover(); p != nil {
			_ = s.transaction.Rollback()
			logrus.WithField("panic", p).Error("transaction rolled back due to panic")
			panic(p)
		}

		if err != nil {
			errRollback := s.transaction.Rollback()
			if errRollback != nil {
				logrus.WithFields(logrus.Fields{
					"err":          err,
					"err_rollback": errRollback,
				}).Errorf("%s[ransaction.Rollback] Error during transaction", logTag)
			}
			return
		}

		errCommit := s.transaction.Commit()
		if errCommit != nil {
			logrus.WithFields(logrus.Fields{
				"err":        err,
				"err_commit": errCommit,
			}).Errorf("%s[ransaction.Commit] Error during transaction", logTag)
		}
	}()

	participantsTx := s.receiptParticipantsRepo.NewTx(tx)
	contactsTx := s.participantContactsRepo.NewTx(tx)

	for _, participant := range participants {
		id, err := participantsTx.InsertMany(ctx, receiptId, []entity.ReceiptParticipant{participant})
		if err != nil {
			return hApperror.InternalServerError(hApperror.AppErrorOpt{
				Message: fmt.Sprintf("%s[participantsTx.InsertMany] Failed to insert participant: %v [receipt_id: %v]", logTag, err, receiptId),
			})
		}

		allConntacts := make([]entity.ParticipantContact, 0, 10)

		for i, participant := range participants {
			for _, contact := range participant.Contacts {
				contact.ParticipantId = id[i]
				allConntacts = append(allConntacts, contact)
			}
		}

		err = contactsTx.InsertMany(ctx, allConntacts)
		if err != nil {
			return hApperror.InternalServerError(hApperror.AppErrorOpt{
				Message: fmt.Sprintf("%s[contactsTx.InsertMany] Failed to insert contacts: %v [receipt_id: %v]", logTag, err, receiptId),
			})
		}
	}

	return nil
}

func (s *receiptParticipant) AddParticipants(ctx context.Context, receiptId int64, participants []entity.ReceiptParticipant) error {
	logTag := s.logTag + "[AddParticipant]"

	deviceId := ctx.Value(hAppconstant.DeviceIdKey).(string)

	receipt, err := s.receiptsRepo.GetByReceiptId(ctx, receiptId, deviceId)
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptsRepo.GetByReceiptId] Failed to get receipt: %v [receipt_id: %v]", logTag, err, receiptId),
		})
	}
	if receipt == nil {
		return hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:            http.StatusNotFound,
			ResponseMessage: "Receipt not found",
		})
	}

	tx, err := s.transaction.Begin()
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[transaction.Begin] Failed to begin transaction: %v [receipt_id: %v]", logTag, err, receiptId),
		})
	}

	defer func() {
		if p := recover(); p != nil {
			_ = s.transaction.Rollback()
			logrus.WithField("panic", p).Error("transaction rolled back due to panic")
			panic(p)
		}

		if err != nil {
			errRollback := s.transaction.Rollback()
			if errRollback != nil {
				logrus.WithFields(logrus.Fields{
					"error":          err,
					"error_rollback": errRollback,
				}).Errorf("%s[ransaction.Rollback] Error during transaction", logTag)
			}
			return
		}

		errCommit := s.transaction.Commit()
		if errCommit != nil {
			logrus.WithFields(logrus.Fields{
				"error":        err,
				"error_commit": errCommit,
			}).Errorf("%s[ransaction.Commit] Error during transaction", logTag)
		}
	}()

	participantsTx := s.receiptParticipantsRepo.NewTx(tx)
	contactsTx := s.participantContactsRepo.NewTx(tx)

	for i := range participants {
		if participants[i].NoticeInterval == 0 {
			participants[i].NoticeInterval = hEntity.Duration(24 * time.Hour)
		}
	}

	partipantIds, err := participantsTx.InsertMany(ctx, receiptId, participants)
	if err != nil {
		if !errors.Is(err, repository.ErrUniqueViolation) {
			return hApperror.InternalServerError(hApperror.AppErrorOpt{
				Message: fmt.Sprintf("%s[participantsTx.InsertMany] Failed to insert participants: %v [receipt_id: %v]", logTag, err, receiptId),
			})
		}

		for _, participant := range participants {
			id, err := participantsTx.InsertMany(ctx, receiptId, []entity.ReceiptParticipant{participant})
			if err != nil {
				return hApperror.InternalServerError(hApperror.AppErrorOpt{
					Message: fmt.Sprintf("%s[oneByOne][participantsTx.InsertMany] Failed to insert participant: %v [receipt_id: %v]", logTag, err, receiptId),
				})
			}

			participantContacts := make([]entity.ParticipantContact, 0, 10)

			for _, contact := range participant.Contacts {
				contact.ParticipantId = id[0]
				participantContacts = append(participantContacts, contact)
			}

			err = contactsTx.InsertMany(ctx, participantContacts)
			if err != nil {
				return hApperror.InternalServerError(hApperror.AppErrorOpt{
					Message: fmt.Sprintf("%s[oneByOne][contactsTx.InsertMany] Failed to insert contacts: %v [receipt_id: %v]", logTag, err, receiptId),
				})
			}
		}

		return nil
	}

	fmt.Println(partipantIds)
	fmt.Println(participants)

	allConntacts := make([]entity.ParticipantContact, 0, 10*len(participants))

	for i, participant := range participants {
		for _, contact := range participant.Contacts {
			contact.ParticipantId = partipantIds[i]
			allConntacts = append(allConntacts, contact)
		}
	}

	err = contactsTx.InsertMany(ctx, allConntacts)
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[contactsTx.InsertMany] Failed to insert contacts: %v [receipt_id: %v]", logTag, err, receiptId),
		})
	}

	return nil
}

func (s *receiptParticipant) GetByReceiptId(ctx context.Context, receiptId int64) ([]entity.ReceiptParticipant, error) {
	logTag := s.logTag + "[GetByReceiptId]"

	deviceId := ctx.Value(hAppconstant.DeviceIdKey).(string)

	receipt, err := s.receiptsRepo.GetByReceiptId(ctx, receiptId, deviceId)
	if err != nil {
		return []entity.ReceiptParticipant{}, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptsRepo.GetByReceiptId] Failed to get receipt: %v [receipt_id: %v]", logTag, err, receiptId),
		})
	}
	if receipt == nil {
		return []entity.ReceiptParticipant{}, hApperror.BadRequestError(hApperror.AppErrorOpt{
			Code:            http.StatusNotFound,
			ResponseMessage: "Receipt not found",
		})
	}

	participants, err := s.receiptParticipantsRepo.GetByReceiptId(ctx, receiptId)
	if err != nil {
		return []entity.ReceiptParticipant{}, hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptParticipantsRepo.GetByReceiptId] Failed to get participants: %v [receipt_id: %v]", logTag, err, receiptId),
		})
	}

	for i := range participants {
		contacts, err := s.participantContactsRepo.GetByParticipantId(ctx, participants[i].ParticipantId)
		if err != nil {
			return participants, hApperror.InternalServerError(hApperror.AppErrorOpt{
				Message: fmt.Sprintf("%s[participantContactsRepo.GetByParticipantId] Failed to get contacts: %v [receipt_id: %v][participant_id: %v]", logTag, err, receiptId, participants[i].ParticipantId),
			})
		}

		participants[i].Contacts = contacts
	}

	return participants, nil
}
