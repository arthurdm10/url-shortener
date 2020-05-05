package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"xojoc.pw/useragent"

	"go.mongodb.org/mongo-driver/mongo"

	netUrl "net/url"

	"github.com/arthurdm10/url-shortener/models"
	"github.com/gofiber/fiber"
	"go.mongodb.org/mongo-driver/bson"
)

type FiberHandler = func(*fiber.Ctx)

func Home(s *Server) FiberHandler {
	return func(c *fiber.Ctx) {
		session := s.sessions.Start(c)
		session.Save(c, session)
		c.Render("./public/index.html", nil)
	}
}

//Shorten shortens a url
func Shorten(s *Server) FiberHandler {
	return func(c *fiber.Ctx) {
		session := s.sessions.Start(c)

		originalURL := c.Body("url-input")
		if originalURL != "" {
			parsedURL, err := netUrl.Parse(originalURL)

			if err != nil {
				log.Println(err.Error())
				c.SendStatus(http.StatusBadRequest)
				return
			}

			if len(parsedURL.Scheme) == 0 {
				parsedURL.Scheme = "http"
			}

			deleteAfterStr := c.Body("deleteAfter")
			var deleteAfter *int
			if deleteAfterStr != "" {
				val, err := strconv.Atoi(deleteAfterStr)
				if err != nil {
					log.Println(err.Error())
					c.SendStatus(http.StatusBadRequest)
					return
				}
				if val > 0 {
					deleteAfter = &val
				}
			}

			url := models.Url{
				Original:    parsedURL.String(),
				DeleteAfter: deleteAfter,
				Session:     session.ID(),
				CreatedAt:   time.Now(),
				Stats:       bson.M{},
			}

			url.Shorten()
			err = s.URLrepo.CreateURL(url)

			if err != nil {
				log.Println(err)
				c.SendStatus(http.StatusBadRequest)
				return
			}

			log.Println(url.Short)
			c.Redirect("/info/" + url.Code)
		}
	}
}

// Redirect redirects to the original url, if its not expired
func Redirect(s *Server) FiberHandler {
	return func(c *fiber.Ctx) {
		//TODO: ignore requests from the same IP/Session
		short := c.Params("short")

		url, err := s.URLrepo.GetUrl(short)

		if err != nil {
			log.Println(err)
			c.SendStatus(http.StatusNotFound)
			return
		}

		if url.Expired() {
			go func(u models.Url) {
				if err = s.URLrepo.DeleteURL(url.Code, url.Session); err != nil {
					log.Printf("Failed to delete url '%s' -- error: %s\n", url.Short, err.Error())
				}

				if err = s.URLRequestRepo.DeleteByURL(url.Short); err != nil {
					log.Printf("Failed to delete urls '%s' requests -- error: %s\n", url.Short, err.Error())
				}
			}(url)

			log.Println("expired")
			c.Redirect("/", http.StatusPermanentRedirect)
			return
		}

		//get information about the client's IP
		go func(clientIP, referer, userAgentStr string) {
			ipInfo := getIPinfo(clientIP)
			userAgent := parseUserAgent(userAgentStr)
			referer = strings.Replace(referer, ".", "_", -1)

			if referer == "" {
				referer = "none"
			}

			err := s.URLrepo.IncrementURLStats(short, strings.ToLower(ipInfo["country"].(string)), referer, userAgent.Name, userAgent.OS)

			if err != nil {
				log.Println("Failed to create URLRequest " + err.Error())
			}

		}(c.IP(), c.Get("referer"), c.Get("user-agent"))

		c.Redirect(url.Original, http.StatusTemporaryRedirect)
	}
}

//URLInfo returns information about the url
func URLInfo(s *Server) FiberHandler {
	return func(c *fiber.Ctx) {

		url, err := s.URLrepo.GetUrlByCode(c.Params("code"))

		if err != nil {
			log.Println(err.Error())
			return
		}

		err = c.Render("./public/info.html", fiber.Map{"url": url})
		if err != nil {
			log.Println(err.Error())
			c.SendStatus(http.StatusInternalServerError)
			return
		}
	}
}

//DeleteURL deletes a url from the database
func DeleteURL(s *Server) FiberHandler {
	return func(c *fiber.Ctx) {
		code := c.Params("code")
		session := s.sessions.Start(c)

		err := s.URLrepo.DeleteURL(code, session.ID())

		if err == mongo.ErrNoDocuments {
			log.Println(err)
			c.Redirect("/myUrls?err=URL not found!")
			return
		} else if err != nil {
			log.Println(err)
			c.SendStatus(http.StatusInternalServerError)
			c.Redirect("/myUrls?err=Internal error")
			return
		}

		c.Redirect("/myUrls?success=deleted")
	}
}

//MyURLs returns urls of the current session
func MyURLs(s *Server) FiberHandler {
	return func(c *fiber.Ctx) {
		session := s.sessions.Start(c)
		urls, err := s.URLrepo.GetURLSBySession(session.ID(), 25)

		if err != nil && err != mongo.ErrNoDocuments {
			log.Println(err.Error())
			c.SendStatus(http.StatusInternalServerError)
			return
		}

		if err = c.Render("./public/myUrls.html", fiber.Map{"urls": urls}); err != nil {
			log.Println(err.Error())
		}
	}
}

func getIPinfo(ip string) bson.M {
	var ipData bson.M
	resp, _ := http.Get("https://extreme-ip-lookup.com/json/" /*+ ip*/)

	json.NewDecoder(resp.Body).Decode(&ipData)
	return ipData
}

func parseUserAgent(userAgent string) *useragent.UserAgent {
	uagent := useragent.Parse(userAgent)

	if uagent != nil {
		return uagent
	}

	return &useragent.UserAgent{
		Name: "",
		OS:   "",
	}
}
