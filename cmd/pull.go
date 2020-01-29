package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// RepoList is a list of images from /v2/_catalog
type RepoList struct {
	Repositories []string `json:"repositories"`
}

// Image is a spec of images and their respective tags
type Image struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func repoList(acrRepo string, acrRefreshToken string) []string {
	var repoList RepoList
	acrAccessToken := genACRAccessToken(acrRefreshToken, acrRepo, "registry:catalog:*")
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://" + acrRepo + "/v2/_catalog", nil)
	if err != nil {
		log.Fatal(err, "Failed to prepare request to /v2/_catalog")
	}
	req.Header.Add("Authorization", "Bearer " + acrAccessToken)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err, "Failed to get /v2/_catalog")
	} else if resp.StatusCode == 401 {
		log.Fatal("Unable to authenticate to ACR.")
	}
	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err, "Failed to read registry list")
	}
	json.Unmarshal(byteValue, &repoList)
	return repoList.Repositories
}

func imageList(acrRepo string, acrRefreshToken string, repos []string) []Image {
	var image Image
	var imageList []Image
	client := &http.Client{}
	for _, repo := range repos {
		acrAccessToken := genACRAccessToken(acrRefreshToken, acrRepo, "repository:" + repo + ":metadata_read")
		req, err := http.NewRequest("GET", "https://" + acrRepo + "/v2/" + repo + "/tags/list", nil)
		if err != nil {
			log.Error(err, "Failed to list tags for %v", repo)
		}	
		req.Header.Add("Authorization", "Bearer " + acrAccessToken)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Failed to list tags for %v\n%v", repo, err)
		} else if resp.StatusCode == 401 {
			log.Fatal("Unable to authenticate to ACR.")
		}
		byteValue, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err, "Failed to read registry list")
		}
		json.Unmarshal(byteValue, &image)
		log.Infof("Image: %v, tags: %v found in src repo.", image.Name, image.Tags)
		imageList = append(imageList, image)
		log.Info(imageList)
	}
	return imageList
}

func imagePull(ctx context.Context, acrRepo string, docker *client.Client, images []Image, authStr string) {
	for _, image := range images {
		for _, tag := range image.Tags {
			imageTag := acrRepo + "/" + image.Name + ":" + tag
			res, err := docker.ImagePull(ctx, imageTag, types.ImagePullOptions{RegistryAuth: authStr})
			if err != nil {
				log.Errorf("Failed to pull image: %v, tag: %v\n %v", image.Name, tag, err)
			}
			defer res.Close()
			log.Info(os.Stdout, res)
		}
	}
}
