package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestValidateName_Valid(t *testing.T) {
	room := &Room{Name: "Conference Room A"}
	err := room.ValidateName()
	assert.NoError(t, err)
}

func TestValidateName_Empty(t *testing.T) {
	room := &Room{Name: ""}
	err := room.ValidateName()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
}

func TestValidateName_TooLong(t *testing.T) {
	room := &Room{Name: "This is a very long room name that exceeds fifty characters limit"}
	err := room.ValidateName()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
}

func TestValidateName_MaxLength(t *testing.T) {
	// Exactly 50 characters
	room := &Room{Name: "12345678901234567890123456789012345678901234567890"}
	err := room.ValidateName()
	assert.NoError(t, err)
}

func TestValidateCode_Valid(t *testing.T) {
	room := &Room{Code: "CONF_A_001"}
	err := room.ValidateCode()
	assert.NoError(t, err)
}

func TestValidateCode_ValidWithHyphen(t *testing.T) {
	room := &Room{Code: "CONF-A-001"}
	err := room.ValidateCode()
	assert.NoError(t, err)
}

func TestValidateCode_ValidWithNumbers(t *testing.T) {
	room := &Room{Code: "Room123"}
	err := room.ValidateCode()
	assert.NoError(t, err)
}

func TestValidateCode_Empty(t *testing.T) {
	room := &Room{Code: ""}
	err := room.ValidateCode()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
}

func TestValidateCode_InvalidCharacters_Space(t *testing.T) {
	room := &Room{Code: "CONF A"}
	err := room.ValidateCode()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
}

func TestValidateCode_InvalidCharacters_Special(t *testing.T) {
	room := &Room{Code: "CONF@A#1"}
	err := room.ValidateCode()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
}

func TestValidateCode_InvalidCharacters_Dot(t *testing.T) {
	room := &Room{Code: "CONF.A.1"}
	err := room.ValidateCode()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
}

func TestValidateCode_TooLong(t *testing.T) {
	room := &Room{Code: "This_is_a_very_long_room_code_that_exceeds_fifty_characters_limit"}
	err := room.ValidateCode()
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidInput, err)
}

func TestValidateCode_MaxLength(t *testing.T) {
	// Exactly 50 characters
	room := &Room{Code: "12345678901234567890123456789012345678901234567890"}
	err := room.ValidateCode()
	assert.NoError(t, err)
}

func TestValidate_Success(t *testing.T) {
	room := &Room{
		Name: "Conference Room A",
		Code: "CONF_A_001",
	}
	err := room.Validate()
	assert.NoError(t, err)
}

func TestValidate_InvalidName(t *testing.T) {
	room := &Room{
		Name: "",
		Code: "CONF_A_001",
	}
	err := room.Validate()
	assert.Error(t, err)
}

func TestValidate_InvalidCode(t *testing.T) {
	room := &Room{
		Name: "Conference Room A",
		Code: "CONF@A",
	}
	err := room.Validate()
	assert.Error(t, err)
}

func TestValidate_BothInvalid(t *testing.T) {
	room := &Room{
		Name: "",
		Code: "CONF@A",
	}
	err := room.Validate()
	assert.Error(t, err)
}

func TestRoom_Structure(t *testing.T) {
	userID := primitive.NewObjectID()
	room := &Room{
		ID:        primitive.NewObjectID(),
		Name:      "Test Room",
		Code:      "TEST_001",
		CreatedBy: userID,
	}

	assert.NotNil(t, room.ID)
	assert.Equal(t, "Test Room", room.Name)
	assert.Equal(t, "TEST_001", room.Code)
	assert.Equal(t, userID, room.CreatedBy)
}
