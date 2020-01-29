package cmd

import (
	"github.com/docker/docker/client"
	"os"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

var (
	rootCmd = &cobra.Command{
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
	rootCmd.Flags().BoolP("push", "p", false, "Push all images and tags")
	viper.BindPFlag("push", rootCmd.Flags().Lookup("push"))
	return rootCmd.Execute()
}

func main(cmd *cobra.Command, args []string) {
	// Prepare authentication tokens
	accessToken, refreshToken, tenantID := parseAzureConfig()
	srcAcrRefreshToken := genACRRefreshToken(accessToken, refreshToken, args[0], tenantID)
	// Grab list of repos, and tags
	repoList := repoList(args[0], srcAcrRefreshToken)
	images := imageList(args[0], srcAcrRefreshToken, repoList)
	log.Info(images)
	os.Exit(0)
	// Create context and docker client
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err, "Failed to create docker client. Is docker running?")
	}
	if viper.GetBool("push") {
		// Pull all images
		srcDockerAuth := genDockerAuth(srcAcrRefreshToken)
		imagePull(ctx, args[0], docker, images, srcDockerAuth)
		// Push all images
		// destAcrRefreshToken := genACRRefreshToken(accessToken, refreshToken, args[1], tenantID)
		// destDockerAuth := genDockerAuth(destAcrRefreshToken)
		// imagePush(ctx, docker, images, destDockerAuth)
	} else {
		log.Info("Not pulling and pushing images. Specify --push to kick off the sync")
	}
}
