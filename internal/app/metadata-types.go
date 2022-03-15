package app

type Metadata struct {
	Sections []Section `json:"sections"`
}

type Section struct {
	Id          string `json:"id"`
	ContentFile string `json:"content-file"`
}
