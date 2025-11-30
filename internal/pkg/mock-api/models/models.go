package models

// ErrorResponse структура ошибки
type ErrorResponse struct {
	Error  string   `json:"error"`
	Fields []string `json:"fields,omitempty"`
}

// MkdirResponse структура ответа для mkdir
type MkdirResponse struct {
	Attributes struct {
		Actor   string `json:"actor"`
		Grantor string `json:"grantor"`
		Mandate string `json:"mandate"`
	} `json:"attributes"`
	Counts struct {
		Files   int `json:"files"`
		Folders int `json:"folders"`
	} `json:"counts"`
	DownloadLimit struct {
		Left      int `json:"left"`
		NextReset int `json:"next_reset"`
		Total     int `json:"total"`
	} `json:"download_limit"`
	Downloads int `json:"downloads"`
	Flags     struct {
		Blocked    bool `json:"blocked"`
		Depo       bool `json:"depo"`
		Favorite   bool `json:"favorite"`
		Restricted bool `json:"restricted"`
	} `json:"flags"`
	Hidden  bool     `json:"hidden"`
	Kind    string   `json:"kind"`
	Link    Link     `json:"link"`
	List    []string `json:"list"`
	Malware struct {
		Status string `json:"status"`
	} `json:"malware"`
	Mtime  int    `json:"mtime"`
	Name   string `json:"name"`
	NodeID string `json:"nodeid"`
	Path   string `json:"path"`
	Size   int    `json:"size"`
	Thumb  struct {
		Xm0  string `json:"xm0"`
		Xms0 string `json:"xms0"`
		Xms4 string `json:"xms4"`
	} `json:"thumb"`
	Type  string `json:"type"`
	Views int    `json:"views"`
}

// AddRequest структура запроса для add
type AddRequest struct {
	Hash    string `json:"hash"`
	LastMod int    `json:"last_modified"`
	Options struct {
		AutoRename bool `json:"autorename"`
		ExclByHash bool `json:"excl_by_hash"`
		Overwrite  bool `json:"overwrite"`
	} `json:"options"`
	Overwrite       bool   `json:"overwrite"`
	Path            string `json:"path"`
	Size            int    `json:"size"`
	UnlimitedUpload bool   `json:"unlimited_upload"`
	UploadType      string `json:"upload_type"`
}

// Link структура для share/unshare ответов
type Link struct {
	Ctime     int    `json:"ctime"`
	Downloads int    `json:"downloads"`
	Expires   int    `json:"expires"`
	ExtID     string `json:"extid"`
	Flags     struct {
		SEOIndexed      bool `json:"SEO_INDEXED"`
		Commentable     bool `json:"commentable"`
		Domestic        bool `json:"domestic"`
		EmailListAccess bool `json:"email_list_access"`
		Writable        bool `json:"writable"`
	} `json:"flags"`
	ID      string `json:"id"`
	Mode    string `json:"mode"`
	Name    string `json:"name"`
	Owner   bool   `json:"owner"`
	Type    string `json:"type"`
	Unknown bool   `json:"unknown"`
	URL     string `json:"url"`
	Views   int    `json:"views"`
}
