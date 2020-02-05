package cmd

import (
	"encoding/json"
	"io"
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
	req, err := http.NewRequest("GET", "https://"+acrRepo+"/v2/_catalog", nil)
	if err != nil {
		log.Fatal(err, "Failed to prepare request to /v2/_catalog")
	}
	req.Header.Add("Authorization", "Bearer "+acrAccessToken)
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

func repoTags(acrRepo string, acrRefreshToken string, repo string) Image {
	var image Image
	client := &http.Client{}
	acrAccessToken := genACRAccessToken(acrRefreshToken, acrRepo, "repository:"+repo+":metadata_read")
	req, err := http.NewRequest("GET", "https://"+acrRepo+"/v2/"+repo+"/tags/list", nil)
	if err != nil {
		log.Error(err, "Failed to list tags for %v", repo)
	}
	req.Header.Add("Authorization", "Bearer "+acrAccessToken)
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
	if len(image.Tags) > 0 {
		log.Debugf("Found image: %v with tags: %v in repo: %v", image.Name, image.Tags, acrRepo)
	} else {
		log.Debugf("Did not find image: %v in repo: %v", repo, acrRepo)
	}
	
	return image
}

func imagePull(ctx context.Context, acrRepo string, docker *client.Client, image Image, authStr string) {
	for _, tag := range image.Tags {
		imageTag := acrRepo + "/" + image.Name + ":" + tag
		log.Infof("Pulling image: %v", imageTag)
		out, err := docker.ImagePull(ctx, imageTag, types.ImagePullOptions{RegistryAuth: authStr})
		if err != nil {
			log.Errorf("Failed to pull image: %v, tag: %v\n %v", image.Name, tag, err)
		}
		defer out.Close()
		io.Copy(os.Stdout, out)
	}
}

func imagePush(ctx context.Context, oldAcrRepo string, acrRepo string, docker *client.Client, image Image, authStr string) {
	for _, tag := range image.Tags {
		oldImageTag := oldAcrRepo + "/" + image.Name + ":" + tag
		imageTag := acrRepo + "/" + image.Name + ":" + tag
		err := docker.ImageTag(ctx, oldImageTag, imageTag)
		if err != nil {
			log.Errorf("Failed to re-tag image: %v with new tag: %v", oldImageTag, imageTag)
		}
		log.Infof("Pushing image: %v", imageTag)
		out, err := docker.ImagePush(ctx, imageTag, types.ImagePushOptions{RegistryAuth: authStr})
		if err != nil {
			log.Errorf("Failed to push image: %v, tag: %v\n %v", image.Name, tag, err)
		}
		defer out.Close()
		io.Copy(os.Stdout, out)
	}
}