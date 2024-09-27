package info

type JumpitPost struct {
	Company     string   `json:"company"`
	Name        string   `json:"name"`
	Skills      []string `json:"skills"`
	Description string   `json:"description"`
	Link        string   `json:"link"`
}

type JumpitDetail struct {
	Tags            string `json:"tags"`
	Congratulations string `json:"congratulations"`
}
