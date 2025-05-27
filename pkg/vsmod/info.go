package vsmod

type ID string

type Version string

type Info struct {
	Type         string         `json:"type"`
	ModID        ID             `json:"modid"`
	Name         string         `json:"name"`
	Authors      []string       `json:"authors"`
	Translators  []string       `json:"translators"`
	Contributors []string       `json:"contributors"`
	Description  string         `json:"description"`
	Version      Version        `json:"version"`
	Dependencies map[ID]Version `json:"dependencies"`
}
