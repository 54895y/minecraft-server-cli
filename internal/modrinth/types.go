package modrinth

type SearchResult struct {
	ProjectID   string   `json:"project_id"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ProjectType string   `json:"project_type"`
	Categories  []string `json:"categories"`
	Downloads   int64    `json:"downloads"`
}

type Version struct {
	ID            string   `json:"id"`
	ProjectID     string   `json:"project_id"`
	Name          string   `json:"name"`
	VersionNumber string   `json:"version_number"`
	DatePublished string   `json:"date_published"`
	Files         []File   `json:"files"`
	GameVersions  []string `json:"game_versions"`
	Loaders       []string `json:"loaders"`
}

type File struct {
	Hashes   map[string]string `json:"hashes"`
	URL      string            `json:"url"`
	Filename string            `json:"filename"`
	Primary  bool              `json:"primary"`
	Size     int64             `json:"size"`
	FileType string            `json:"file_type"`
}
