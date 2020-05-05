package main

import (
	"time"

	repo "github.com/arthurdm10/url-shortener/repository"
	"github.com/gofiber/fiber"
	"github.com/gofiber/logger"
	"github.com/gofiber/session"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	app            *fiber.Fiber
	URLrepo        repo.URLRepository
	URLRequestRepo repo.URLRequestRepository
	port           string
	sessions       *session.Session
}

// NewServer creates a new instance
func NewServer(port string, db *mongo.Database) Server {
	sv := Server{
		URLrepo:        repo.URLRepository{DB: db},
		URLRequestRepo: repo.URLRequestRepository{DB: db},
		port:           port,
		sessions: session.New(session.Config{
			Expires: time.Hour * 3600,
		}),
	}
	app := fiber.New(
		&fiber.Settings{
			CaseSensitive: true,
		},
	)

	app.Use(logger.New())

	app.Static("/css", "./public/css")
	app.Static("/js", "./public/js")
	app.Static("/info/topojson", "./public/js")

	app.Get("/", Home(&sv))

	app.Get("/info/:code", URLInfo(&sv))
	app.Get("/myUrls", MyURLs(&sv))
	app.Post("/shorten", Shorten(&sv))

	app.Get("/del/:code", DeleteURL(&sv))

	app.Get("/:short", Redirect(&sv))

	sv.app = app

	return sv
}

//Run run server
func (s *Server) Run() {
	s.app.Listen(s.port)
}
