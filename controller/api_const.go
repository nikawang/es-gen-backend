package controller

const (
	// API represents the group of API.
	API = "/api"
	// APIBooks represents the group of book management API.
	APIES     = API + "/es"
	APIESNAME = API + "/es" + "/:name/:namespace"
	// APIBooksID represents the API to get book data using id.
	// APIBooksID = APIBooks + "/:id"
	// APICategories represents the group of category management API.
	APICategories = API + "/categories"
	// APIFormats represents the group of format management API.
	APIFormats = API + "/formats"
)

const (
	// APIAccount represents the group of auth management API.
	APIAccount = API + "/auth"
	// APIAccountLoginStatus represents the API to get the status of logged in account.
	APIAccountLoginStatus = APIAccount + "/loginStatus"
	// APIAccountLoginAccount represents the API to get the logged in account.
	APIAccountLoginAccount = APIAccount + "/loginAccount"
	// APIAccountLogin represents the API to login by session authentication.
	APIAccountLogin = APIAccount + "/login"
	// APIAccountLogout represents the API to logout.
	APIAccountLogout = APIAccount + "/logout"
)

const (
	// APIHealth represents the API to get the status of this application.
	APIHealth = API + "/health"
)
