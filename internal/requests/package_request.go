package requests

type UploadPackageRequest struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DistTags    struct {
		Latest string `json:"latest"`
	} `json:"dist-tags"`
	Versions map[string]struct {
		Name        string            `json:"name"`
		Version     string            `json:"version"`
		Description string            `json:"description"`
		Main        string            `json:"main"`
		Scripts     map[string]string `json:"scripts"`
		Author      interface{}       `json:"author"`
		License     string            `json:"license"`
		Readme      string            `json:"readme"`
		ID          string            `json:"_id"`
		NodeVersion string            `json:"_nodeVersion"`
		NpmVersion  string            `json:"_npmVersion"`
		Dist        struct {
			Integrity string `json:"integrity"`
			Shasum    string `json:"shasum"`
			Tarball   string `json:"tarball"`
		} `json:"dist"`
	} `json:"versions"`
	Access      *string `json:"access"`
	Attachments map[string]struct {
		ContentType string `json:"content_type"`
		Data        string `json:"data"`
		Length      int    `json:"length"`
	} `json:"_attachments"`
}
