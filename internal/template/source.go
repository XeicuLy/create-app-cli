package template

type Template struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Manifest struct {
	Templates []Template `json:"templates"`
}

type TemplateSource interface {
	ListTemplates() ([]Template, error)
	Fetch(id, destDir string) error
}
