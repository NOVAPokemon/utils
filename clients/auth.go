package clients

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NOVAPokemon/utils"
	"github.com/NOVAPokemon/utils/api"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type AuthClient struct {
	Jar *cookiejar.Jar
}

func (client *AuthClient) LoginWithUsernameAndPassword(username, password string) error {

	httpClient := &http.Client{
		Jar: client.Jar,
	}

	jsonStr, err := json.Marshal(utils.UserJSON{Username: username, Password: password})
	if err != nil {
		log.Error(err)
	}

	host := fmt.Sprintf("%s:%d", utils.Host, utils.AuthenticationPort)
	loginUrl := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   api.LoginPath,
	}

	req, err := http.NewRequest("POST", loginUrl.String(), bytes.NewBuffer(jsonStr))

	if err != nil {
		log.Error(err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)

	if err != nil {
		log.Error(err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("unexpected reponse %d", resp.StatusCode))
	}

	return nil
}

func (client *AuthClient) GetInitialTokens(username string) error {
	httpClient := &http.Client{
		Jar: client.Jar,
	}

	host := fmt.Sprintf("%s:%d", utils.Host, utils.TrainersPort)
	generateTokensUrl := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   fmt.Sprintf(api.GenerateAllTokensPath, username),
	}

	log.Info("requesting tokens at ", generateTokensUrl.String())

	req, err := http.NewRequest("GET", generateTokensUrl.String(), nil)

	if err != nil {
		log.Error(err)
		return err
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		log.Error(err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("unexpected reponse %s", resp.Status))
	}

	return nil
}

func (client *AuthClient) Register(username string, password string) error {

	httpClient := &http.Client{}

	jsonStr, err := json.Marshal(utils.UserJSON{Username: username, Password: password})
	if err != nil {
		log.Error(err)
	}

	host := fmt.Sprintf("%s:%d", utils.Host, utils.AuthenticationPort)
	loginUrl := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   api.RegisterPath,
	}

	req, err := http.NewRequest("POST", loginUrl.String(), bytes.NewBuffer(jsonStr))

	if err != nil {
		log.Error(err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req)

	if err != nil {
		log.Error(err)
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("unexpected reponse %d", res.StatusCode))
	}

	return nil
}
