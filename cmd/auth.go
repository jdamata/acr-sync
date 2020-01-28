package cmd

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

// https://github.com/Azure/acr/blob/master/docs/AAD-OAuth.md#authenticating-docker-with-an-acr-refresh-token

// AzureConfig struct used to unmarshal ~/.azure/accessTokens.json
type AzureConfig []struct {
	ExpiresIn        int    `json:"expiresIn"`
	Authority        string `json:"_authority"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
	ExpiresOn        string `json:"expiresOn"`
	UserID           string `json:"userId"`
	IsMRRT           bool   `json:"isMRRT"`
	ClientID         string `json:"_clientId"`
	Resource         string `json:"resource"`
	AccessToken      string `json:"accessToken"`
	IdentityProvider string `json:"identityProvider,omitempty"`
	Oid              string `json:"oid,omitempty"`
}

// ACRToken retrieved from
type ACRToken struct {
	Token string `json:"refresh_token"`
}

func parseAzureConfig() (string, string, string) {
	var accessToken string
	var refreshToken string
	var tenantID string
	var expiresOn string
	var azureConfig AzureConfig
	usr, _ := user.Current()
	homedir := usr.HomeDir
	authenticationFile, err := os.Open(homedir + "/.azure/accessTokens.json")
	if err != nil {
		log.Fatalf("Failed to open ~/.azure/accessTokens.json. Be sure to run az login first! %v\n", err)
	}
	byteValue, _ := ioutil.ReadAll(authenticationFile)
	json.Unmarshal(byteValue, &azureConfig)
	for i := 0; i < len(azureConfig); i++ {
		if azureConfig[i].ExpiresOn > expiresOn {
			expiresOn = azureConfig[i].ExpiresOn
			accessToken = azureConfig[i].AccessToken
			refreshToken = azureConfig[i].RefreshToken
			tenantID = strings.TrimPrefix(azureConfig[i].Authority, "https://login.microsoftonline.com/")
		}
	}
	return accessToken, refreshToken, tenantID
}

func genACRToken(accessToken string, refreshToken string, acrRepo string, tenantID string) ACRToken {
	var acrToken ACRToken
	data := url.Values{}
	data.Set("grant_type", "access_token_refresh_token")
	data.Set("access_token", accessToken)
	data.Set("refresh_token", refreshToken)
	data.Set("service", acrRepo)
	data.Set("tenant", tenantID)
	resp, err := http.PostForm("https://"+acrRepo+"/oauth2/exchange", data)
	if err != nil {
		log.Fatal(err, "Authenticate via the cli with az login first")
	}
	byteValue, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(byteValue, &acrToken)
	return acrToken
}

func genDockerAuth(acrToken ACRToken) string {
	authConfig := types.AuthConfig{
		Username: "00000000-0000-0000-0000-000000000000",
		Password: acrToken.Token,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		log.Fatal(err)
	}
	return base64.URLEncoding.EncodeToString(encodedJSON)
}
