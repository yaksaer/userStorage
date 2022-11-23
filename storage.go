package userStorage

import (
	"fmt"
	"github.com/gin-contrib/authz"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"sync"
)

// userGrade struct for user information
type userGrade struct {
	UserId        string `json:"user_id" validate:"required"`
	PostpaidLimit int    `json:"postpaid_limit"`
	Spp           int    `json:"spp"`
	ShippingFee   int    `json:"shipping_fee"`
	ReturnFee     int    `json:"return_fee"`
}

// UserStore is the main struct, it contains userGrade struct, mutex, setUser and getUser.
// Create an instance of UserStore, by using New()
type UserStore struct {
	sync.Mutex

	users    map[string]userGrade
	lastUser string
}

// New create UserStore struct
func New() *UserStore {
	return &UserStore{
		users: make(map[string]userGrade),
	}
}

// checkFullness check received data on fullness
func checkFullness(user userGrade) bool {
	if user.UserId == "" || user.PostpaidLimit == -1 || user.Spp == -1 ||
		user.ShippingFee == -1 || user.ReturnFee == -1 {
		return false
	}
	return true
}

// SetUser receive JSON and save it in userGrade struct. Use it in POST(set user) request as handler
func (us *UserStore) SetUser(c *gin.Context) {
	// Use if you receive UserID in each request
	us.Lock()
	defer us.Unlock()
	user := userGrade{}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	us.users[user.UserId] = user
	c.JSON(http.StatusOK, gin.H{"Id": user.UserId})

	// Use if you don't receive UserID in each request, but you're receiving data about one user until it won't be complete

	/*us.Lock()
	defer us.Unlock()
	user := userGrade{}
	if us.lastUser == "" {
		user = userGrade{
			UserId:        "",
			PostpaidLimit: -1,
			Spp:           -1,
			ShippingFee:   -1,
			ReturnFee:     -1,
		}
	} else {
		user = us.users[us.lastUser]
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if !checkFullness(user) {
		us.lastUser = user.UserId
		us.users[user.UserId] = user
		c.JSON(http.StatusOK, gin.H{"Id": user.UserId})
		return
	}
	us.lastUser = ""
	us.users[user.UserId] = user
	c.JSON(http.StatusOK, gin.H{"Id": user.UserId})
	*/
}

// GetUser find user by userID and return information ad JSON. Use it in GET(get user) request as handler
func (us *UserStore) GetUser(c *gin.Context) {
	us.Lock()
	defer us.Unlock()
	userId := c.DefaultQuery("user_id", "")
	answer, ok := us.users[userId]
	if !ok {
		c.JSON(http.StatusBadRequest, "Заданный UserID не найден")
		return
	}
	c.JSON(http.StatusOK, answer)
}

// AuthRequired check user's authorization
func AuthRequired(c *gin.Context) {
	a := authz.BasicAuthorizer{}
	user := a.GetUserName(c.Request)
	if user != os.Getenv("USER") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.Next()
}

// StartSet start server and wait for user information. You should start it as goroutine
func (us *UserStore) StartSet(port string) {
	router := gin.Default()
	auth := router.Group("/")
	auth.Use(AuthRequired)
	{
		auth.POST("/", us.SetUser)
	}
	fmt.Println("Starting server on", port)
	err := router.Run(port)
	if err != nil {
		return
	}
}

// StartGet start server and wait url with userID. You should start it as goroutine
func (us *UserStore) StartGet(port string) {
	router := gin.Default()

	router.GET("/", us.GetUser)
	fmt.Println("Starting server on", port)
	err := router.Run(port)
	if err != nil {
		return
	}
}
