package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/digitalocean/godo"
	"os"
	"strconv"
	"strings"
)

func getEnvVarPresent(name string) string {
	key, present := os.LookupEnv(name)
	if !present {
		fmt.Printf("%s not present as environment variable\n", name)
		os.Exit(1)
	}
	return key
}

func main() {
	doApiToken := getEnvVarPresent("DO_API_TOKEN")
	clusterId := getEnvVarPresent("K8S_CLUSTER_ID")
	spacesRegion := getEnvVarPresent("SPACES_REGION")
	spacesAccessKey := getEnvVarPresent("SPACES_KEY_ACCESS")
	spacesSecretKey := getEnvVarPresent("SPACES_SECRET_KEY")
	spacesBucket := getEnvVarPresent("SPACES_BUCKET")
	spacesFilename := getEnvVarPresent("SPACES_FILENAME")
	kubernetesMaxExpiry , err:= strconv.ParseInt(getEnvVarPresent("K8S_MAX_EXPIRY_TOKEN"), 10, 64)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// generate a new kubeconfig token
	client := godo.NewFromToken(doApiToken)
	ctx := context.TODO()
	kubeconfig, _, err := client.Kubernetes.GetKubeConfigWithExpiry(ctx, clusterId, kubernetesMaxExpiry)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// push new kubeconfig to spaces bucket
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(spacesAccessKey, spacesSecretKey, ""),
		Endpoint:    aws.String("https://" + spacesRegion + ".digitaloceanspaces.com"),
		Region:      aws.String("us-east-1"),
	}

	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)

	object := s3.PutObjectInput{
		Bucket: aws.String(spacesBucket),
		Key:    aws.String(spacesFilename),
		Body:   strings.NewReader(string(kubeconfig.KubeconfigYAML)),
		ACL:    aws.String("private"),
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Successfully generated a new kubeconfig file")
	fmt.Printf("Token expiration: %ds\nFilename: %s\nBucket: %s\nRegion: %s\n", kubernetesMaxExpiry, spacesFilename, spacesBucket, spacesRegion)
}