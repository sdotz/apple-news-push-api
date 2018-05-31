# Apple News Push API

A library and command line tool for interacting with [the Apple News push API](https://developer.apple.com/library/content/documentation/General/Conceptual/News_API_Ref/index.html#//apple_ref/doc/uid/TP40015409-CH2-SW1)

The library aims to be complete and easy to use both as tool for interacting with the API manually, and for building larger applications which syndicate content to the Apple News platform.

## Todo

- [X] Implement all API methods and options
- [X] CLI
- [X] Push notifications
- [X] Scan article.json for bundle:// URLs, find those files in the bundle, and add to multipart upload
- [ ] Refactor all API methods to return response objects and move JSON printing out to CLI land
- [ ] Add ability to download a bundle locally (also w/ font files stored somewhere)
- [ ] Make CLI friendly with validations and meaningful errors
- [ ] Add tests
- [ ] Clean up project structure, dirs, package names etc.
- [ ] Implement paging searchArticles through channels or iterator
- [ ] Interactive paging through searchArticles in CLI


## Issues

I think the promoteArticles API might be broken when sending an empty array. Docs say you can do this but all I get back is a server error.


## Usage

Install and use as a command line tool:

`go install github.com/sdotz/apple-news-push-api/cmd/anews`

Set the envoriment variables `APPLE_NEWS_API_KEY`, `APPLE_NEWS_API_SECRET`, `CHANNEL_ID`

`anews --help`