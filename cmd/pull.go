package cmd

import (
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

func imageList(ctx context.Context, docker client.Client) {
	var imageList []types.ImageSummary
	imageList, err := docker.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		log.Fatal(err, "Can't list images in ACR repo")
	}
	for _, image := range imageList {
		// imageInspect, _, _ := docker.ImageInspectWithRaw(ctx, image.ID)
		fmt.Println(image.RepoTags)
	}
}

func imagePull(ctx context.Context, docker client.Client, images []string, authStr string) {
	for _, image := range images {
		image, err := docker.ImagePull(ctx, image, types.ImagePullOptions{RegistryAuth: authStr})
		if err != nil {
			log.Error(err, "Failed to pull image")
		}
		defer image.Close()
		log.Info(os.Stdout, image)
	}
}
