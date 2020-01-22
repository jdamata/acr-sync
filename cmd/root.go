package cmd

import (
	"os"
	"os/user"
	"io/ioutil"
	"encoding/base64"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// AccessToken struct used to unmarshal ~/.azure/accessTokens.json
type AccessToken []struct {
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

var (
	rootCmd  = &cobra.Command{
		Use:   "acrsync",
		Short: "Sync images from one ACR to another",
		Args:  cobra.ExactArgs(2),
		Run:   main,
	}
)

func logging() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

// Execute executes the root command.
func Execute(version string) error {
	rootCmd.Version = version
	rootCmd.Flags().BoolP("all", "a", false, "Push all images and tags")
	viper.BindPFlag("all", rootCmd.Flags().Lookup("all"))
	return rootCmd.Execute()
}

func grabToken() string {
	var token string
	var expireson string
	usr, _ := user.Current()
	homedir := usr.HomeDir
	accessTokensFile, err := os.Open(homedir + "/.azure/accessTokens.json")
	if err != nil {
		log.Fatalf("Failed to open ~/.azure/accessTokens.json. Be sure to run az login first! %v\n", err)
	}
	byteValue, _ := ioutil.ReadAll(accessTokensFile)
	var accesstoken AccessToken
	json.Unmarshal(byteValue, &accesstoken)
	for i := 0; i < len(accesstoken); i++ {
		if accesstoken[i].ExpiresOn > expireson {
			expireson = accesstoken[i].ExpiresOn
			token = accesstoken[i].AccessToken
		}
	}
	return token
}

func dockerCredential() string {
	token := grabToken()
	authConfig := types.AuthConfig{
		Username: "<token>",
		Password: token,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		log.Fatal(err)
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	return authStr
}

func dockerClient(acrrepo string) *client.Client {
	cli, err := client.NewClientWithOpts(client.WithHost(acrrepo), client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create docker client. %v\n", err)
	}
	return cli
}

func main(cmd *cobra.Command, args []string) {

	// Authenticate to both docker registries
	acrSrcURL := "https://" + args[0] // + "/v2/"
	// acrDestURL := "https://" + args[1]
	srcClient := dockerClient(acrSrcURL)
	// destClient := dockerClient(acrDestURL)

	// Grab docker credential
	// authStr := dockerCredential()
	ctx := context.Background()
	_, err := srcClient.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		log.Fatal(err)
	}
}
