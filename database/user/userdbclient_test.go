package user

import (
	"crypto/rand"
	"github.com/NOVAPokemon/utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var userMockup = utils.User{
	Username:     "user1",
	PasswordHash: randSeq(256),
}

func TestMain(m *testing.M) {
	_ = removeAll()

	res := m.Run()

	_ = removeAll()

	os.Exit(res)
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

	res, err := AddUser(&userMockup)

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	log.Println(res)
}

func TestGetAll(t *testing.T) {
	res, err := GetAllUsers()

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	for i, item := range res {
		log.Println(i, item.Username)
	}
}

func TestGetByID(t *testing.T) {

	user, err := GetUserByUsername(userMockup.Username)

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	assert.Equal(t, user.Username, userMockup.Username)
	assert.Equal(t, user.PasswordHash, userMockup.PasswordHash)
}

func TestGetByUsername(t *testing.T) {

	user, err := GetUserByUsername(userMockup.Username)

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

	_, err := UpdateUser(userMockup.Username, &toUpdate)
	if err != nil {
		log.Error(err)
		t.Fail()
	}

	updatedUser, err := GetUserByUsername(userMockup.Username)
	if err != nil {
		log.Error(err)
		t.Fail()
	}

	assert.Equal(t, toUpdate.Username, updatedUser.Username)
	assert.Equal(t, toUpdate.PasswordHash, updatedUser.PasswordHash)
}

func TestDelete(t *testing.T) {
	userMockup.Username = "user2"
	oID, err := AddUser(&userMockup)

	if err != nil {
		log.Println(err)
		t.Fail()
	}

	err = DeleteUser(oID)
	if err != nil {
		log.Error(err)
		t.Fail()
	}

	users, err := GetAllUsers()
	if err != nil {
		log.Error(err)
		t.Fail()
	}

	for _, user := range users {
		if user.Username == userMockup.Username {
			log.Errorf("delete did not delete user %+v", user)
			t.Fail()
		}

	}
}
