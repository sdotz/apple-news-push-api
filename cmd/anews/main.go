package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/apple-news-push-api/pkg/api"
	"gopkg.in/alecthomas/kingpin.v2"
)

const BASE_URL = "https://news-api.apple.com"

var (
	//verbose   = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	channelId = kingpin.Flag("channelId", "The ID of the channel to use").Default(os.Getenv("CHANNEL_ID")).String()
	apiKey    = kingpin.Flag("apiKey", "The API key to use when calling the API").Default(os.Getenv("APPLE_NEWS_API_KEY")).String()
	apiSecret = kingpin.Flag("apiSecret", "The API secret to use when calling the API").Default(os.Getenv("APPLE_NEWS_API_SECRET")).String()
	baseUrl   = kingpin.Flag("baseUrl", "The base URL to use for API calls").Default(BASE_URL).String()

	readCommand = kingpin.Command("read", "Read a channel, section or article")
	articleId   = readCommand.Command("article", "Read an article").Arg("Article ID", "The (apple) ID of the article to read").String()
	readChannel = readCommand.Command("channel", "Read a channel")
	sectionId   = readCommand.Command("section", "Read a section").Arg("Section ID", "The ID of the section to read").String()

	listChannelSections = kingpin.Command("list", "List sections")

	searchCommand  = kingpin.Command("search", "List articles in a channel or section")
	searchOptions  = newSearchOptions(searchCommand)
	searchFromDate = searchCommand.Flag("fromDate", "Start paging from this date (formatted like 2006-01-02)").String()
	searchToDate   = searchCommand.Flag("toDate", "End paging at this date (formatted like 2006-01-02)").String()

	createCommand = kingpin.Command("create", "Create an article")
	bundlePath    = createCommand.Arg("bundlePath", "Path to the bundle. It should contain article.json and any images that are referenced within it").Required().ExistingFileOrDir()
	createOptions = newCreateUpdateOptions(createCommand)

	updateCommand    = kingpin.Command("update", "Update an article")
	updateBundlePath = updateCommand.Flag("bundlePath", "Path to the bundle. It should contain article.json and any images that are referenced within it").ExistingFileOrDir()
	revision         = updateCommand.Arg("revision ID", "The revision ID of the article to update").Required().String()
	updateArticleId  = updateCommand.Arg("article ID", "The (apple) ID of the article to update").Required().String()
	updateOptions    = newCreateUpdateOptions(updateCommand)

	promoteCommand    = kingpin.Command("promote", "Promote an article")
	promoteSectionId  = promoteCommand.Arg("section ID", "The section ID to promote articles in").Required().String()
	promoteArticleIds = promoteCommand.Arg("article IDs", "The article IDs to promote. If none, then promoted articles will be removed").Strings()

	deleteCommand   = kingpin.Command("delete", "Delete an article")
	deleteArticleId = deleteCommand.Arg("article ID", "The ID of the article to delete").Required().String()
)

func main() {
	command := kingpin.Parse()

	channelID := *channelId
	key := *apiKey
	secret := *apiSecret
	baseURL := *baseUrl
	articleID := *articleId

	c := api.NewClient(&http.Client{}, baseURL, key, secret, channelID)

	switch command {
	case "read article":
		resp, err := c.ReadArticle(articleID)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if j, err := json.Marshal(resp); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(string(j))
		}
	case "read channel":
		resp, err := c.ReadChannel(channelID)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if j, err := json.Marshal(resp); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		} else {
			fmt.Println(string(j))
		}
	case "read section":
		sectionID := *sectionId
		resp, err := c.ReadSection(sectionID)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if j, err := json.Marshal(resp); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		} else {
			fmt.Println(string(j))
		}
	case "list":
		resp, err := c.ListSections()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if j, err := json.Marshal(resp); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		} else {
			fmt.Println(string(j))
		}
	case "search":
		if from, err := time.Parse("2006-01-02", *searchFromDate); err == nil {
			searchOptions.FromDate = &from
		}
		if to, err := time.Parse("2006-01-02", *searchToDate); err == nil {
			searchOptions.ToDate = &to
		}
		resp, err := c.SearchArticles(searchOptions)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if j, err := json.Marshal(resp); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		} else {
			fmt.Println(string(j))
		}
	case "create":
		f, err := os.Open(*bundlePath)
		if err != nil {
			log.Fatalf("%v", err)
		}
		defer f.Close()
		c.CreateArticle(f, nil)
	case "update":
		if len(*updateBundlePath) > 0 {
			f, err := os.Open(*bundlePath)
			if err != nil {
				log.Fatalf("%v", err)
			}
			defer f.Close()

			c.UpdateArticle(articleID, f, updateOptions)
		} else {
			c.UpdateArticleMetadata(articleID, updateOptions)
		}
	case "promote":
		c.PromoteArticles(*promoteSectionId, *promoteArticleIds)
	case "delete":
		c.DeleteArticle(*deleteArticleId)
	}

}

func newCreateUpdateOptions(cmd *kingpin.CmdClause) *api.Metadata {
	options := &api.Metadata{}
	cmd.Flag("sections", "The sections the article should appear in").StringsVar(&options.Links.Sections)
	cmd.Flag("isSponsored", "Marks the article as sponsored").BoolVar(&options.IsSponsored)
	cmd.Flag("isPreview", "Sets the article to preview mode").BoolVar(&options.IsPreview)
	cmd.Flag("accessoryText", "Sets text below the article excerpt in channel view. Default is the author").StringVar(&options.AccessoryText)
	cmd.Flag("maturityRating", "Sets the article's maturity rating").HintOptions(api.MaturityRatingKids, api.MaturityRatingMature, api.MaturityRatingGeneral).EnumVar(&options.MaturityRating, api.MaturityRatingKids, api.MaturityRatingMature, api.MaturityRatingGeneral)
	cmd.Flag("isCandidateToBeFeatured", "Sets the article as a candidate to be featured").BoolVar(&options.IsCandidateToBeFeatured)
	cmd.Flag("isHidden", "Sets the article to hidden").BoolVar(&options.IsHidden)
	return options
}

func newSearchOptions(cmd *kingpin.CmdClause) *api.SearchArticlesOptions {
	defaultSearchOpts := api.DefaultSearchArticlesOptions()
	cmd.Flag("pageSize", "The amount of articles per page to return").IntVar(&defaultSearchOpts.PageSize)
	cmd.Flag("sortDir", "Direction to sort by date").HintOptions(api.SORTDIR_ASC, api.SORTDIR_DESC).EnumVar(&defaultSearchOpts.SortDir, api.SORTDIR_ASC, api.SORTDIR_DESC)
	return defaultSearchOpts
}
