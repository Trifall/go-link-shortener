package lib

type Routes struct {
	Localhost    string
	API          string
	Health       string
	V1           string
	Keys         keysRoutes
	Links        linksRoutes
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

type linksRoutes struct {
	Base     string
	Shorten  string
	Retrieve string
	Delete   string
	Update   string
}

var ROUTES = Routes{
	Localhost: "http://localhost:8080",
	API:       "/api",
	Health:    "/health",
	V1:        "/v1",
	Keys: keysRoutes{
		Base:     "/keys",
		Validate: "/validate",
		Generate: "/generate",
		Update:   "/update",
		Delete:   "/delete",
	},
	Links: linksRoutes{
		Base:     "/links",
		Shorten:  "/shorten",
		Retrieve: "/retrieve",
		Delete:   "/delete",
		Update:   "/update",
	},
	Docs:         "/docs",
	DocsJsonFile: "/docs/doc.json",
}

type ReservedRoutes struct {
	API  string
	Docs string
}

var RESERVED_ROUTES = ReservedRoutes{
	API:  "api",
	Docs: "docs",
}
