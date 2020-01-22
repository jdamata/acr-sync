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

// ACRRefreshToken retrieved from /oauth2/exchange
type ACRRefreshToken struct {
	Token string `json:"refresh_token"`
}

// ACRAccessToken retrieved from /oauth2/token
type ACRAccessToken struct {
	Token string `json:"access_token"`
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

func genACRRefreshToken(accessToken string, refreshToken string, acrRepo string, tenantID string) string {
	var ACRRefreshToken ACRRefreshToken
	data := url.Values{}
	data.Set("grant_type", "access_token_refresh_token")
	data.Set("access_token", accessToken)
	data.Set("refresh_token", refreshToken)
	data.Set("service", acrRepo)
	data.Set("tenant", tenantID)
	resp, err := http.PostForm("https://" + acrRepo + "/oauth2/exchange", data)
	if err != nil {
		log.Fatal(err, "Authenticate via the cli with az login first")
	} else if resp.StatusCode == 401 {
		log.Fatal("Unable to authenticate to /oauth2/exchange. Authenticate via the cli with az login first")
	}
	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(byteValue, &ACRRefreshToken)
	log.Info("Generated a /oauth2/exchange refresh_token successfully")
	return ACRRefreshToken.Token
}

func genACRAccessToken(ACRRefreshToken string, acrRepo string, scope string) string {
	var ACRAccessToken ACRAccessToken
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", ACRRefreshToken)
	data.Set("service", acrRepo)
	data.Set("scope", scope)
	resp, err := http.PostForm("https://" + acrRepo + "/oauth2/token", data)
	if err != nil {
		log.Fatal(err, "Authenticate via the cli with az login first")
	} else if resp.StatusCode == 401 {
		log.Fatal("Unable to authenticate to /oauth2/token")
	}
	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(byteValue, &ACRAccessToken)
	return ACRAccessToken.Token
}

func genDockerAuth(AcrRefreshToken string) string {
	authConfig := types.AuthConfig{
		Username: "00000000-0000-0000-0000-000000000000",
		Password: AcrRefreshToken,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		log.Fatal(err)
	}
	return base64.URLEncoding.EncodeToString(encodedJSON)
}
