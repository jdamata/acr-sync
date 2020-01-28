package cmd

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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
	rootCmd.Flags().BoolP("all", "a", false, "Push all images and tags")
	viper.BindPFlag("all", rootCmd.Flags().Lookup("all"))
	return rootCmd.Execute()
}

func main(cmd *cobra.Command, args []string) {
	//accessToken, refreshToken, tenantID := parseAzureConfig()
	//acrToken := genACRToken(accessToken, refreshToken, args[0], tenantID)
	//authStr := genDockerAuth(acrToken)
	docker, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err, "Can't start docker client")
	}
	ctx := context.Background()
	imageList(ctx, docker)
}
