package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"github.com/sdotz/apple-news-push-api/pkg"
	"log"
)

var (
	verbose   = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	channelId = kingpin.Flag("channelId", "The ID of the channel to use").Default(os.Getenv("CHANNEL_ID")).String()
	apiKey    = kingpin.Flag("apiKey", "The API key to use when calling the API").Default(os.Getenv("APPLE_NEWS_API_KEY")).String()
	apiSecret = kingpin.Flag("apiSecret", "The API secret to use when calling the API").Default(os.Getenv("APPLE_NEWS_API_SECRET")).String()
	baseUrl   = kingpin.Flag("baseUrl", "The base URL to use for API calls").Default(os.Getenv("BASE_URL")).String()

	readCommand = kingpin.Command("read", "Read a channel, section or article")
	articleId   = readCommand.Command("article", "Read an article").Arg("Article ID", "The (apple) ID of the article to read").String()
	readChannel = readCommand.Command("channel", "Read a channel")
	sectionId   = readCommand.Command("section", "Read a section").Arg("Section ID", "The ID of the section to read").String()

	listChannelSections = kingpin.Command("list", "List sections")

	searchCommand = kingpin.Command("search", "List articles in a channel or section")

	createCommand = kingpin.Command("create", "Create an article")
	bundlePath    = createCommand.Arg("bundlePath", "Path to the bundle. It should contain article.json and any necessary images.").Required().ExistingFileOrDir()

	updateCommand  = kingpin.Command("update", "Update an article")
	promoteCommand = kingpin.Command("promote", "Promote an article")
	deleteCommand  = kingpin.Command("delete", "Delete an article")
)

func main() {
	command := kingpin.Parse()

	switch command {
	case "read article":
		pkg.ReadArticle(*articleId, *apiKey, *apiSecret, *baseUrl)
	case "read channel":
		pkg.ReadChannel(*channelId, *apiKey, *apiSecret, *baseUrl)
	case "read section":
		pkg.ReadSection(*sectionId, *apiKey, *apiSecret, *baseUrl)
	case "list":
		pkg.ListSections(*channelId, *apiKey, *apiSecret, *baseUrl)
	case "search":
		searchOpts := pkg.DefaultSearchArticlesOptions()
		pkg.SearchArticles(*channelId, searchOpts, *apiKey, *apiSecret, *baseUrl)
	case "create":
		f, err := os.Open(*bundlePath)
		if err != nil {
			log.Fatal(err.Error())
		}
		defer f.Close()
		pkg.CreateArticle(*channelId, f, nil, *apiKey, *apiSecret, *baseUrl)
	case "update":
	case "promote":
	case "delete":

	}

}
