package repository

import (
	"context"
	"log"
	"time"

	"github.com/arthurdm10/url-shortener/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type URLRepository struct {
	DB *mongo.Database
}

// CreateURL creates a URL in the database
func (u *URLRepository) CreateURL(url models.Url) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := u.DB.Collection("urls").InsertOne(ctx, url)

	return err
}

// DeleteURL deletes a URL from the database
func (u *URLRepository) DeleteURL(code, session string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	res := u.DB.Collection("urls").FindOneAndDelete(ctx, bson.M{"code": code, "session": session})

	return res.Err()
}

func (u *URLRepository) GetUrl(short string) (models.Url, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	log.Println(short)
	urlDoc := u.DB.Collection("urls").FindOne(ctx, bson.M{"short_url": short})

	if urlDoc.Err() != nil {
		return models.Url{}, urlDoc.Err()
	}

	url := models.Url{}

	err := urlDoc.Decode(&url)

	if err != nil {
		return models.Url{}, err
	}

	return url, nil
}

//GetURLSBySession get urls from a session
func (u *URLRepository) GetURLSBySession(session string, limit int) ([]bson.M, error) {
	cursor, err := u.DB.Collection("urls").Aggregate(context.TODO(), bson.A{
		bson.M{
			"$match": bson.M{
				"session": bson.M{
					"$eq": session,
				},
			},
		},
		bson.M{"$limit": limit},
		bson.M{
			"$lookup": bson.M{
				"from":         "requests",
				"localField":   "short_url",
				"foreignField": "url_short",
				"as":           "reqs",
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":          0,
				"original_url": 1,
				"short_url":    1,
				"code":         1,
				"created_at":   1,
				"totalRequests": bson.M{
					"$size": "$reqs",
				},
			},
		},
		bson.M{"$sort": bson.M{"created_at": 0}},
	})

	if err != nil {
		return nil, err
	}

	var urls []bson.M

	err = cursor.All(context.Background(), &urls)

	return urls, err
}

func (u *URLRepository) GetUrlRequests(code string) (bson.M, error) {
	cursor, err := u.DB.Collection("urls").Aggregate(context.TODO(), bson.A{
		bson.M{"$match": bson.M{"code": code}},
		bson.M{
			"$lookup": bson.M{
				"from": "requests",
				"let": bson.M{
					"url_code": "$code",
				},
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							"$expr": bson.M{
								"$eq": bson.A{
									"$$url_code", code,
								},
							},
						},
					},
					bson.M{"$limit": 25},
				},
				"as": "reqs",
			},
		},
	})

	if err != nil {
		return nil, err
	}

	cursor.Next(context.Background())
	var urlInfo bson.M
	err = cursor.Decode(&urlInfo)

	if err != nil {
		return nil, err
	}
	countries := make(map[string]int)
	referers := make(map[string]int)
	browsers := make(map[string]int)

	for _, request := range urlInfo["reqs"].(bson.A) {
		req := request.(bson.M)
		ipReq := req["ip"].(bson.M)

		countryCode := ipReq["country"].(string)
		referer := req["referer"].(string)

		if countryCode != "" {
			_, found := countries[countryCode]
			if found {
				countries[countryCode] += 1
				continue
			}

			countries[countryCode] = 1
		}

		if referer != "" {
			_, found := referers[referer]
			if found {
				referers[referer] += 1
				continue
			}

			referers[referer] = 1
		}
		if req["user_agent"] == nil {
			continue
		}
		userAgent := req["user_agent"].(bson.M)

		if browser, found := userAgent["browser"].(string); found {
			if browser != "" {
				_, found := browsers[browser]
				if found {
					browsers[browser] += 1
					continue
				}

				browsers[browser] = 1
			}
		}
	}
	urlInfo["countries"] = countries
	urlInfo["referers"] = referers
	urlInfo["browsers"] = browsers
	return urlInfo, err
}
