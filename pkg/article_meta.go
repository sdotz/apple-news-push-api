package pkg

const (
	CoverArtTypeImage = "image"
)

type CoverArt struct {
	Type                 CoverArtType `json:"type"`
	URL                  string       `json:"URL"`
	AccessibilityCaption string       `json:"accessibilityCaption,omitempty"`
}

type ArticleMetadata struct {
	Authors             []string               `json:"authors,omitempty"`
	CampaignData        map[string]interface{} `json:"campaignData,omitempty"`
	CanonicalURL        string                 `json:"canonicalURL,omitempty"`
	CoverArts           []CoverArt             `json:"coverArt,omitempty"`
	DateCreated         string                 `json:"dateCreated,omitempty"`
	DateModified        string                 `json:"dateModified,omitempty"`
	DatePublished       string                 `json:"datePublished,omitempty"`
	Excerpt             string                 `json:"excerpt,omitempty"`
	GeneratorIdentifier string                 `json:"generatorIdentifier,omitempty"`
	GeneratorName       string                 `json:"generatorName,omitempty"`
	GeneratorVersion    string                 `json:"generatorVersion,omitempty"`
	Keywords            []string               `json:"keywords,omitempty"`
	ThumbnailURL        string                 `json:"thumbnailURL,omitempty"`
	TransparentToolbar  string                 `json:"transparentToolbar,omitempty"`
	VideoURL            string                 `json:"videoURL,omitempty"`
}
