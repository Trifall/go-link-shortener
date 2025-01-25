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
	NotFound     string
}

type keysRoutes struct {
	Base        string
	RetrieveAll string
	Validate    string
	Generate    string
	Update      string
	Delete      string
}

type linksRoutes struct {
	Base             string
	RetrieveAll      string
	RetrieveAllByKey string
	Shorten          string
	Retrieve         string
	Delete           string
	Update           string
}

var ROUTES = Routes{
	Localhost: "http://localhost",
	API:       "/api",
	Health:    "/health",
	V1:        "/v1",
	Keys: keysRoutes{
		Base:        "/keys",
		RetrieveAll: "/retrieve-all",
		Validate:    "/validate",
		Generate:    "/generate",
		Update:      "/update",
		Delete:      "/delete",
	},
	Links: linksRoutes{
		Base:             "/links",
		RetrieveAll:      "/retrieve-all",
		RetrieveAllByKey: "/retrieve-all-by-key",
		Shorten:          "/shorten",
		Retrieve:         "/retrieve",
		Delete:           "/delete",
		Update:           "/update",
	},
	Docs:         "/docs",
	DocsJsonFile: "/docs/doc.json",
	NotFound:     "/404",
}

type ReservedRoutes struct {
	API      string
	Docs     string
	NotFound string
}

var RESERVED_ROUTES = ReservedRoutes{
	API:      "api",
	Docs:     "docs",
	NotFound: "404",
}
