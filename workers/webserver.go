package workers

import (
	"go-link-shortener/api"
	"go-link-shortener/lib"
	"go-link-shortener/utils"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func InitializeWebserver(env *utils.Env) error {
	log.Println("⏳ Initializing API...")
	// Create a new chi router
	r := chi.NewRouter()

	// TODO: might need to update this to allow for the frontend to connect
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Middleware stack
	r.Use(middleware.Logger)    // Log API requests
	r.Use(middleware.Recoverer) // Recover from panics without crashing server

	r.Mount(lib.ROUTES.API, api.InitializeAPIRouter())

	log.Println("✔️  API initialized successfully.")

	if env.ENABLE_DOCS == "true" {

		log.Println("⏳ Setting up swagger API docs...")

		r.Get(lib.ROUTES.Docs+"/*", httpSwagger.Handler(
			httpSwagger.URL(lib.ROUTES.Localhost+":"+env.SERVER_PORT+lib.ROUTES.DocsJsonFile),
			httpSwagger.AfterScript(craftPostScript()),
			httpSwagger.UIConfig(map[string]string{
				"deepLinking":     "true",
				"filter":          "false",
				"showExtensions":  "true",
				"syntaxHighlight": `{"active":"true"}`,
			}),
		))

		log.Println("✔️  Swagger API docs set up successfully.")
	} else {
		log.Println("⚠️  Swagger API docs are disabled. To enable them, set ENABLE_DOCS=true in your .env file.")
	}

	log.Println("⏳ Setting up redirect router...")
	r.Mount("/", api.RedirectRouter())
	log.Println("✔️  Redirect router set up successfully.")

	portString := ":" + env.SERVER_PORT
	log.Println("✔️  Starting server on port " + portString)

	// Start the server
	if err := http.ListenAndServe(portString, r); err != nil {
		return err
	}

	return nil
}

func craftPostScript() string {
	return `const topbarElement = document.querySelector('.topbar'); if (topbarElement) {
			topbarElement.remove();
		}

		// Add blue link styling to all <a> tags with the class "link"
		const linkElements = document.querySelectorAll('a.link');
		linkElements.forEach(link => {
			link.style.color = '#007BFF'; // Blue color
			link.style.textDecoration = 'underline'; // Underline
		});

		// Function to check for the presence of the info__contact div and add the <h1> element
		const checkAndAddH1 = () => {
				const infoContactDiv = document.querySelector('div.info__contact');
				if (infoContactDiv) {
					const h1Element = document.createElement('h1');
					h1Element.textContent = 'Contact/Support';
					infoContactDiv.insertBefore(h1Element, infoContactDiv.firstChild);

					const childs = infoContactDiv.querySelectorAll('div > a');
					for(const child of childs) {
						child.style.fontSize = "1.5rem";
					}

					return true; // Successfully added the <h1> element
				}
			return false; // Element not found yet
		};

		const fixPadding = () => {
			// Add margin-bottom of 1rem to all <li> elements inside div.markdown > ul
			const liElements = document.querySelectorAll('div.markdown > ul li');
			if(liElements.length > 0) {
				liElements.forEach(li => {
					li.style.marginBottom = '1rem';
				});
				return true; // Successfully added margin-bottom
			}
			return false; // Element not found yet
		}

		// Function to check for the presence of the <h1> with content "Authentication" and modify the <p> tag
		const checkAndModifyP = () => {
			const h1Elements = document.querySelectorAll('h1');
			for (const h1 of h1Elements) {
				if (h1.textContent === 'Authentication') {
					const pElement = h1.nextElementSibling;
					if (pElement && pElement.tagName === 'P') {
						pElement.style.lineHeight = '2';
						return true; // Successfully modified the <p> element
					}
				}
			}
			return false; // Element not found yet
		}

		const fixTitle = () => {
			const fixedTitle = "Jerren's Link Shortener API - Docs"
			if(document && document.title !== fixedTitle){
				document.title = fixedTitle;
				return true;
			}
			return false;
		}

		// Loop every 100ms until the info__contact div is found, the <h1> is added, and the <p> is modified
		(function() {
			return new Promise((resolve, reject) => {
				const interval = setInterval(() => {
					if (checkAndAddH1() && fixPadding() && checkAndModifyP() && fixTitle()) {
						clearInterval(interval);
						resolve();
					}
				}, 10);
			});
		})();`
}
