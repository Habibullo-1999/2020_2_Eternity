package usecase

import (
	"errors"
	"github.com/go-park-mail-ru/2020_2_Eternity/configs/config"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/domain"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/notifications"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/pin"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/user"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	NoteComment = iota
	NotePin
	NoteFollow
)

type Usecase struct {
	noteRepo  notifications.IRepository
	pinRepo pin.IRepository
	userRepo user.IRepository
}

func NewUsecase(nr notifications.IRepository, pr pin.IRepository, ur user.IRepository) *Usecase {
	return &Usecase{
		noteRepo:  nr,
		pinRepo: pr,
		userRepo: ur,
	}
}

func (uc *Usecase) getDstNoteComment (n *domain.NoteComment) ([]int, error) {
	p, err := uc.pinRepo.GetPin(n.PinId)
	if err != nil {
		config.Lg("notifications_usecase", "getDstNoteComment").Error(err.Error())
		return nil, err
	}
	toUsers := []int{p.UserId}

	return toUsers, nil
}

func (uc *Usecase) getDstNotePin (p *domain.NotePin) ([]int, error) {
	toUsers, err := uc.userRepo.GetFollowersIds(p.UserId)
	if err != nil {
		config.Lg("notifications_usecase", "getDstNotePin").Error(err.Error())
		return nil, err
	}

	return toUsers, nil
}

func (uc *Usecase) getDstNoteFollow (f *domain.NoteFollow) ([]int, error) {
	return []int{f.UserId}, nil
}

func (uc *Usecase) CreateNotes(iNote interface{}) error {
	var noteType int
	var toUsers []int
	var err error

	switch note := iNote.(type) {
	case domain.NoteComment:
		noteType = NoteComment
		toUsers, err = uc.getDstNoteComment(&note)
	case domain.NotePin:
		noteType = NotePin
		toUsers, err = uc.getDstNotePin(&note)
	case domain.NoteFollow:
		noteType = NoteFollow
		toUsers, err = uc.getDstNoteFollow(&note)
	default:
		config.Lg("notifications_usecase", "CreateNote").Error("Unknown notification type")
		return errors.New("Unknown notification type")
	}

	if err != nil {
		config.Lg("notifications_usecase", "CreateNote").Error(err.Error())
		return err
	}

	encoded, err := bson.Marshal(iNote)
	if err != nil {
		config.Lg("notifications_usecase", "CreateNote").Error(err.Error())
		return err
	}


	note := domain.Notification{
		Type: noteType,
		EncodedData: encoded,
	}
	for _, id := range toUsers {
		note.ToUserId = id
		err = uc.noteRepo.StoreNote(&note)

		if err != nil {
			config.Lg("notifications_usecase", "CreateNote").Error(err.Error())
			return err
		}
	}

	return nil
}


func (uc *Usecase) GetUserNotes(userId int) ([]domain.NoteResp, error) {
	modelNotes, err :=  uc.noteRepo.GetNotesToUser(userId)

	respNotes := []domain.NoteResp{}
	for _, mNote := range modelNotes {
		respNotes = append(respNotes, domain.NoteResp{
			Id: mNote.Id,
			Type: mNote.Type,
			EncodedData: mNote.EncodedData,
			CreationTime: mNote.CreationTime,
			IsRead: mNote.IsRead,
		})
	}

	return respNotes, err
}

func (uc *Usecase) UpdateNote(noteId int) error {
	return uc.noteRepo.UpdateNoteIsRead(noteId)
}

func (uc *Usecase) UpdateUserNotes(userId int) error {
	return uc.noteRepo.UpdateUserNotes(userId)
}
