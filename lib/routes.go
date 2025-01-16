package lib

type Routes struct {
	API     string
	Health  string
	V1      string
	Shorten string
	Keys    keysRoutes
}

type keysRoutes struct {
	Base     string
	Validate string
	Generate string
	Update   string
	Delete   string
}

var ROUTES = Routes{
	API:     "/api",
	Health:  "/health",
	V1:      "/v1",
	Shorten: "/shorten",
	Keys: keysRoutes{
		Base:     "/keys",
		Validate: "/validate",
		Generate: "/generate",
		Update:   "/update",
		Delete:   "/delete",
	},
}
