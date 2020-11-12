package usecase

import (
	"errors"
	"github.com/go-park-mail-ru/2020_2_Eternity/configs/config"
	"github.com/go-park-mail-ru/2020_2_Eternity/pkg/domain"
	mock_notifications "github.com/go-park-mail-ru/2020_2_Eternity/pkg/notifications/mock"
	mock_pin "github.com/go-park-mail-ru/2020_2_Eternity/pkg/pin/mock"
	mock_user "github.com/go-park-mail-ru/2020_2_Eternity/pkg/user/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
	"os"
	"testing"
)

var (
	userId = 321

	toUsers = []int{2}

	noteComment = domain.NoteComment{
		Id: 1,
		Path: []int32{1},
		Content: "content",
		PinId: 1,
		UserId: 1,
	}

	notePin = domain.NotePin{
		Id: 1,
		Title: "title",
		ImgLink: "link",
		UserId: 1,
	}

	noteFollow = domain.NoteFollow{
		FollowerId: 21,
		UserId: 32,
	}

	noteRespMany = []domain.NoteResp {
		{
			Id: 12,
			Type: NoteComment,
			EncodedData: []byte("123123"),
			IsRead: false,
		},
		{
			Id: 13,
			Type: NotePin,
			EncodedData: []byte("1223"),
			IsRead: false,
		},
	}
)


func TestMain(m *testing.M) {
	config.Conf = config.NewConfigTst()
	code := m.Run()
	os.Exit(code)
}


func TestCreateNotesSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockNotes := mock_notifications.NewMockIRepository(ctrl)
	mockPin := mock_pin.NewMockIRepository(ctrl)
	mockUser := mock_user.NewMockIRepository(ctrl)

	uc := NewUsecase(mockNotes, mockPin, mockUser)


	// Success comment

	mockPin.EXPECT().
		GetPin(gomock.Eq(noteComment.PinId)).
		Return(domain.Pin{UserId: toUsers[0]}, nil).
		Times(1)

	encoded, err := bson.Marshal(noteComment)
	require.Nil(t, err)
	mockNotes.EXPECT().
		StoreNote(gomock.Eq(&domain.Notification{
			ToUserId: toUsers[0],
			Type: NoteComment,
			EncodedData: encoded,
		})).
		Return(nil).
		Times(1)

	err = uc.CreateNotes(noteComment)
	assert.Nil(t, err)

	// Success pin

	mockUser.EXPECT().
		GetFollowersIds(gomock.Eq(notePin.UserId)).
		Return(toUsers, nil).
		Times(1)

	encoded, err = bson.Marshal(notePin)
	require.Nil(t, err)
	mockNotes.EXPECT().
		StoreNote(gomock.Eq(&domain.Notification{
			ToUserId: toUsers[0],
			Type: NotePin,
			EncodedData: encoded,
		})).
		Return(nil).
		Times(1)

	err = uc.CreateNotes(notePin)
	assert.Nil(t, err)


	// Success follow


	encoded, err = bson.Marshal(noteFollow)
	require.Nil(t, err)
	mockNotes.EXPECT().
		StoreNote(gomock.Eq(&domain.Notification{
			ToUserId: noteFollow.UserId,
			Type: NoteFollow,
			EncodedData: encoded,
		})).
		Return(nil).
		Times(1)

	err = uc.CreateNotes(noteFollow)
	assert.Nil(t, err)

}


func TestCreateNotesFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockNotes := mock_notifications.NewMockIRepository(ctrl)
	mockPin := mock_pin.NewMockIRepository(ctrl)
	mockUser := mock_user.NewMockIRepository(ctrl)

	uc := NewUsecase(mockNotes, mockPin, mockUser)

	// wrong note type

	err := uc.CreateNotes(1)
	assert.NotNil(t, err)

	// comment error

	mockPin.EXPECT().
		GetPin(gomock.Eq(noteComment.PinId)).
		Return(domain.Pin{UserId: toUsers[0]}, errors.New("")).
		Times(1)


	err = uc.CreateNotes(noteComment)
	assert.NotNil(t, err)

	// store note error

	mockPin.EXPECT().
		GetPin(gomock.Eq(noteComment.PinId)).
		Return(domain.Pin{UserId: toUsers[0]}, nil).
		Times(1)

	encoded, err := bson.Marshal(noteComment)
	require.Nil(t, err)
	mockNotes.EXPECT().
		StoreNote(gomock.Eq(&domain.Notification{
			ToUserId: toUsers[0],
			Type: NoteComment,
			EncodedData: encoded,
		})).
		Return(errors.New("")).
		Times(1)

	err = uc.CreateNotes(noteComment)
	assert.NotNil(t, err)
}


func TestGetUserNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockNotes := mock_notifications.NewMockIRepository(ctrl)
	mockPin := mock_pin.NewMockIRepository(ctrl)
	mockUser := mock_user.NewMockIRepository(ctrl)

	uc := NewUsecase(mockNotes, mockPin, mockUser)

	// Success

	mockNotes.EXPECT().
		GetNotesToUser(gomock.Eq(userId)).
		DoAndReturn(func(uId int) ([]domain.Notification, error) {
			mNotes := []domain.Notification{}
			for _, rNote := range noteRespMany {
				mNotes = append(mNotes, domain.Notification{
					Id: rNote.Id,
					Type: rNote.Type,
					EncodedData: rNote.EncodedData,
					CreationTime: rNote.CreationTime,
					IsRead: rNote.IsRead,
				})
			}
			return mNotes, nil
		}).
		Times(1)


	nsResp, err := uc.GetUserNotes(userId)
	assert.Nil(t, err)
	assert.Equal(t, noteRespMany, nsResp)
}