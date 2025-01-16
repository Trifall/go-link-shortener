package lib

type Routes struct {
	Localhost    string
	API          string
	Health       string
	V1           string
	Shorten      string
	Keys         keysRoutes
	Docs         string
	DocsJsonFile string
}

type keysRoutes struct {
	Base     string
	Validate string
	Generate string
	Update   string
	Delete   string
}

var ROUTES = Routes{
	Localhost: "http://localhost:8080",
	API:       "/api",
	Health:    "/health",
	V1:        "/v1",
	Shorten:   "/shorten",
	Keys: keysRoutes{
		Base:     "/keys",
		Validate: "/validate",
		Generate: "/generate",
		Update:   "/update",
		Delete:   "/delete",
	},
	Docs:         "/docs/*",
	DocsJsonFile: "/docs/doc.json",
}
