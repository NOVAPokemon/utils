package user

import (
	"crypto/rand"
	"github.com/NOVAPokemon/utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var userMockup = utils.User{
	Username:     "user1",
	PasswordHash: randSeq(256),
}

func randSeq(n int) []byte {
	key := make([]byte, n)

	_, err := rand.Read(key)
	if err != nil {
		log.Error(err)
	}
	return key
}

func TestAddUser(t *testing.T) {

	err, res := AddUser(&userMockup)

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	log.Println(res)
}

func TestGetAll(t *testing.T) {
	res := GetAllUsers()
	for i, item := range res {
		log.Println(i, item.Username)
	}
}

func TestGetByID(t *testing.T) {

	err, user := GetUserById(userMockup.Username)

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	assert.Equal(t, user.Username, userMockup.Username)
	assert.Equal(t, user.PasswordHash, userMockup.PasswordHash)
}

func TestGetByUsername(t *testing.T) {

	err, user := GetUserByUsername(userMockup.Username)

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	assert.Equal(t, user.Username, userMockup.Username)
	assert.Equal(t, user.PasswordHash, userMockup.PasswordHash)
}

func TestUpdate(t *testing.T) {
	toUpdate := utils.User{
		Username:     "Updated_John",
		PasswordHash: randSeq(256),
	}

	err, _ := UpdateUser(userMockup.Username, &toUpdate)

	if err != nil {
		log.Error(err)
		t.Fail()
	}
	err, updatedUser := GetUserById(userMockup.Username)

	assert.Equal(t, toUpdate.Username, updatedUser.Username)
	assert.Equal(t, toUpdate.PasswordHash, updatedUser.PasswordHash)
}

func TestDelete(t *testing.T) {
	userMockup.Username = "user2"
	err, oID := AddUser(&userMockup)

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	err = DeleteUser(oID)

	if err != nil {
		log.Error(err)
		t.Fail()
	}

	users := GetAllUsers()

	for _, user := range users {
		if user.Username == userMockup.Username {
			log.Errorf("delete did not delete user %+v", user)
			t.Fail()
		}

	}
}
