package cmd

import (
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"github.com/docker/docker/api/types/filters"
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
	rootCmd.Flags().BoolP("sync", "s", false, "Sync all images and tags")
	rootCmd.Flags().BoolP("prune", "p", false, "Prune local docker images after pushing")
	viper.BindPFlag("sync", rootCmd.Flags().Lookup("sync"))
	viper.BindPFlag("prune", rootCmd.Flags().Lookup("prune"))
	return rootCmd.Execute()
}

func main(cmd *cobra.Command, args []string) {
	if !viper.GetBool("sync") {
		log.Info("Not pulling and pushing images. Specify --sync to kick off the sync")
	}
	// Prepare authentication tokens
	accessToken, refreshToken, tenantID := parseAzureConfig()
	srcAcrRefreshToken := genACRRefreshToken(accessToken, refreshToken, args[0], tenantID)
	// Grab list of repos
	repoList := repoList(args[0], srcAcrRefreshToken)
	// Create context and docker client
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err, "Failed to create docker client. Is docker running?")
	}
	for _, repo := range repoList {
		// Grab tags for repo
		tags := repoTags(args[0], srcAcrRefreshToken, repo)
		if viper.GetBool("sync") {
			// Pull all images
			srcDockerAuth := genDockerAuth(srcAcrRefreshToken)
			imagePull(ctx, args[0], docker, tags, srcDockerAuth)
			// Push all images
			destAcrRefreshToken := genACRRefreshToken(accessToken, refreshToken, args[1], tenantID)
			destDockerAuth := genDockerAuth(destAcrRefreshToken)
			imagePush(ctx, args[0], args[1], docker, tags, destDockerAuth)
		}
	}
	if viper.GetBool("prune") {
		docker.ImagesPrune(ctx, filters.Args{})
	}
}
