package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"
)

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	token := os.Getenv("GH_TOKEN")
	if t := req.QueryStringParameters["gh_token"]; t != "" {
		token = t
	}

	repos := req.MultiValueQueryStringParameters["repo"]
	if len(repos) == 0 {
		repos = []string{req.QueryStringParameters["repo"]}
	}

	if len(repos) == 0 {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Missing repo parameter",
		}, nil
	}

	codeOwnersFilePath := "CODEOWNERS"
	if p := req.QueryStringParameters["codeowners"]; p != "" {
		codeOwnersFilePath = p
	}

	results, err := query(token, codeOwnersFilePath, repos)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       err.Error(),
		}, nil
	}

	if strings.ToLower(req.QueryStringParameters["format"]) == "json" {
		log.Println("JSON FORMAT!!!")
		data, err := json.Marshal(results)
		if err != nil {
			return &events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       err.Error(),
			}, nil
		}

		return &events.APIGatewayProxyResponse{
			StatusCode:      200,
			Headers:         map[string]string{"Content-Type": "application/json"},
			Body:            string(data),
			IsBase64Encoded: false,
		}, nil
	}

	var output strings.Builder
	for _, r := range results {
		_, _ = output.WriteString(r.Repo + "\n")
		for _, p := range r.Paths {
			_, _ = output.WriteString(fmt.Sprintf("Path %s\n", p.Path))
			for _, o := range p.Owners {
				_, _ = output.WriteString(fmt.Sprintf("  - %s\n", o))
			}
		}
		_, _ = output.WriteString("\n")
	}

	return &events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "text/plain"},
		Body:            output.String(),
		IsBase64Encoded: false,
	}, nil
}

func query(token, codeOwnersFilePath string, repos []string) ([]Result, error) {
	if token == "" {
		return nil, errors.New("gh_token is required")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	var results []Result
	for _, fullrepo := range repos {
		split := strings.Split(fullrepo, "/")
		owner := split[0]
		repo := split[1]

		fileContent, _, resp, err := client.Repositories.GetContents(ctx, owner, repo, codeOwnersFilePath, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting code owners file: %s", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error getting code owners file. GH Request Status code: %s", resp.Status)
		}

		content, err := fileContent.GetContent()
		if err != nil {
			return nil, err
		}

		results = append(results, Result{
			Repo:  fullrepo,
			Paths: parseContent(content),
		})
	}

	return results, nil

}

func main() {
	lambda.Start(handler)
}

type Result struct {
	Repo  string  `json:"repo"`
	Paths []Paths `json:"paths"`
}

type Paths struct {
	Path   string   `json:"path"`
	Owners []string `json:"owners"`
}

func parseContent(content string) []Paths {
	var owners []Paths
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, " ")
		owners = append(owners, Paths{
			Path:   parts[0],
			Owners: parts[1:],
		})
	}

	return owners
}
