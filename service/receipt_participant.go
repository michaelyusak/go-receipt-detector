package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"receipt-detector/constant"
	"receipt-detector/entity"
	"receipt-detector/repository"
	"time"

	hAppconstant "github.com/michaelyusak/go-helper/appconstant"
	hApperror "github.com/michaelyusak/go-helper/apperror"
	hEntity "github.com/michaelyusak/go-helper/entity"
	hHelper "github.com/michaelyusak/go-helper/helper"
	"github.com/sirupsen/logrus"
)

type receiptParticipant struct {
	receiptParticipantsRepo repository.ReceiptParticipants
	participantContactsRepo repository.ParticipantContacts
	receiptsRepo            repository.Receipts

	smtpHelper *hHelper.SmtpHelper

	transaction repository.Transaction

	allowedContactTypes []string

	logTag string
}

type ReceiptParticipantOpt struct {
	ReceiptParticipantsRepo repository.ReceiptParticipants
	ParticipantContactsRepo repository.ParticipantContacts
	ReceiptsRepo            repository.Receipts

	SmptpHelper *hHelper.SmtpHelper

	Transaction repository.Transaction

	AllowedContactTypes []string
}

func NewReceiptParticipant(opt ReceiptParticipantOpt) *receiptParticipant {
	return &receiptParticipant{
		receiptParticipantsRepo: opt.ReceiptParticipantsRepo,
		participantContactsRepo: opt.ParticipantContactsRepo,
		receiptsRepo:            opt.ReceiptsRepo,

		smtpHelper: opt.SmptpHelper,

		transaction: opt.Transaction,

		allowedContactTypes: opt.AllowedContactTypes,

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

func (s *receiptParticipant) NotifyParticipants(logTag string, receiptName string, participants []entity.ReceiptParticipant) {
	for _, p := range participants {
		if !p.Notifying {
			continue
		}

		for _, c := range p.Contacts {
			switch c.ContactType {
			case "email":
				data := map[string]any{
					"name":       p.ParticipantName,
					"bill_title": receiptName,
				}

				req, err := s.smtpHelper.
					NewRequest([]string{c.ContactValue}, constant.ParticipantAddedSubject).
					SetBody(constant.ParticipantAddedTemplate, data)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"err":            err,
						"participant_id": c.ParticipantId,
					}).Errorf("%s[AddParticipants][smtpHelper.NewRequest] Failed to create smtp request", logTag)
				}

				err = req.Send()
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"err":            err,
						"participant_id": c.ParticipantId,
					}).Errorf("%s[AddParticipants][req.Send] Failed to send smtp request", logTag)
				}

				logrus.Infof("%s[AddParticipants][req.Send] Success to send smtp request", logTag)
			}
		}
	}
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
				}).Errorf("%s[transaction.Rollback] Error during transaction", logTag)
			}
			logrus.Infof("%s[transaction.Rollback] transaction rollbacked", logTag)
			return
		}

		errCommit := s.transaction.Commit()
		if errCommit != nil {
			logrus.WithFields(logrus.Fields{
				"error_commit": errCommit,
			}).Errorf("%s[transaction.Commit] Error during transaction", logTag)
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

	go s.NotifyParticipants(logTag, receipt.ReceiptName, participants)

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

func (s *receiptParticipant) GetAllowedContactTypes() []string {
	return s.allowedContactTypes
}

func (s *receiptParticipant) UpdateOne(ctx context.Context, participant entity.ReceiptParticipant) error {
	logTag := s.logTag + "[UpdateOne]"

	existingParticipant, err := s.receiptParticipantsRepo.GetByParticipantId(ctx, participant.ParticipantId)
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptParticipantsRepo.GetByParticipantId] Failed get participant: %v [participant_id: %v]", logTag, err, participant.ParticipantId),
		})
	}
	if existingParticipant == nil {
		return hApperror.BadRequestError(hApperror.AppErrorOpt{
			ResponseMessage: "participant not found",
			Message:         fmt.Sprintf("%s[ParticipantNotFound] Participant not found [participant_id: %v]", logTag, participant.ParticipantId),
		})
	}

	tx, err := s.transaction.Begin()
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[transaction.Begin] Failed to begin transaction: %v [participant_id: %v]", logTag, err, participant.ParticipantId),
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
				}).Errorf("%s[transaction.Rollback] Error during transaction", logTag)
			}
			logrus.Infof("%s[transaction.Rollback] transaction rollbacked", logTag)
			return
		}

		errCommit := s.transaction.Commit()
		if errCommit != nil {
			logrus.WithFields(logrus.Fields{
				"error_commit": errCommit,
			}).Errorf("%s[transaction.Commit] Error during transaction", logTag)
		}
	}()

	receiptParticipantsTx := s.receiptParticipantsRepo.NewTx(tx)
	contactsTx := s.participantContactsRepo.NewTx(tx)

	fmt.Printf("participant: %+v\n", participant)

	err = receiptParticipantsTx.UpdateOne(ctx, participant)
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptParticipantTx.UpdateOne] Failed to update participant: %v [participant_id: %v]", logTag, err, participant.ParticipantId),
		})
	}

	err = contactsTx.DeleteByParticipantId(ctx, participant.ParticipantId)
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[contactsTx.DeleteByParticipantId] Failed to delete old contacts: %v [participant_id: %v]", logTag, err, participant.ParticipantId),
		})
	}

	err = contactsTx.InsertMany(ctx, participant.Contacts)
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[contactsTx.InsertMany] Failed to insert new contacts: %v [participant_id: %v]", logTag, err, participant.ParticipantId),
		})
	}

	go func() {
		if !participant.Notifying {
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		deviceId := ctx.Value(hAppconstant.DeviceIdKey).(string)

		receipt, err := s.receiptsRepo.GetByReceiptId(c, participant.ReceiptId, deviceId)
		if err != nil || receipt == nil {
			logrus.WithFields(logrus.Fields{
				"receipt": receipt,
				"err":     err,
			}).Errorf("%s[receiptsRepo.GetByReceiptId] Failed to get receipt", logTag)
			return
		}

		s.NotifyParticipants(logTag, receipt.ReceiptName, []entity.ReceiptParticipant{participant})
	}()

	return nil
}

func (s *receiptParticipant) DeleteOne(ctx context.Context, participantId int64) error {
	logTag := s.logTag + "[DeleteOne]"

	tx, err := s.transaction.Begin()
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[transaction.Begin] Failed to begin transaction: %v [participant_id: %v]", logTag, err, participantId),
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
				}).Errorf("%s[transaction.Rollback] Error during transaction", logTag)
			}
			logrus.Infof("%s[transaction.Rollback] transaction rollbacked", logTag)
			return
		}

		errCommit := s.transaction.Commit()
		if errCommit != nil {
			logrus.WithFields(logrus.Fields{
				"error_commit": errCommit,
			}).Errorf("%s[transaction.Commit] Error during transaction", logTag)
		}
	}()

	receiptParticipantsTx := s.receiptParticipantsRepo.NewTx(tx)
	participantContactsTx := s.participantContactsRepo.NewTx(tx)

	err = receiptParticipantsTx.DeleteByParticipantIds(ctx, []int64{participantId})
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[receiptParticipantsTx.DeleteByParticipantIds] Failed to delete receipt participant: %v [participant_id: %v]", logTag, err, participantId),
		})
	}

	err = participantContactsTx.DeleteByParticipantId(ctx, participantId)
	if err != nil {
		return hApperror.InternalServerError(hApperror.AppErrorOpt{
			Message: fmt.Sprintf("%s[participantContactsTx.DeleteByParticipantId] Failed to delete participant contacts: %v [participant_id: %v]", logTag, err, participantId),
		})
	}

	return nil
}
