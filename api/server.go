package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	db "github.com/r-scheele/escout/db/sqlc"
	"github.com/r-scheele/escout/token"
	"github.com/r-scheele/escout/util"
	"github.com/robfig/cron/v3"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
	cron       *cron.Cron
	colly      *colly.Collector
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	var domains []string
	for domain := range util.AllowedDomains {
		domains = append(domains, domain)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		cron:       cron.New(),
		colly: colly.NewCollector(
			colly.AllowedDomains(domains...),
		),
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()
	router.GET("/", server.ping)

	router.POST("/users", server.createUser)
	router.GET("/users/:id", server.getUser)
	router.GET("/users", server.getUsers)

	router.POST("/products", server.trackProduct)
	router.GET("/products/:id", server.getProduct)
	router.PUT("/products/:id", server.updateProduct)
	router.GET("/products/:id/prices", server.getProductPriceChanges)
	router.GET("/products", server.getProducts)

	router.POST("/users/login", server.loginUser)
	// router.POST("/tokens/renew_access", server.renewAccessToken)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func (server *Server) ping(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "pong",
	})
}
