package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

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
	urlDoc := u.DB.Collection("urls").FindOne(ctx, bson.M{"short": short},
		&options.FindOneOptions{
			Projection: bson.M{"original": 1, "created_at": 1, "delete_after": 1},
		},
	)

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

//IncrementURLStats increments the total of clicks/requests in this URL
func (u *URLRepository) IncrementURLStats(short, country, referer, browser, os string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	updateQuery := bson.M{
		"stats.clicks":               1,
		"stats.countries." + country: 1,
		"stats.referer." + referer:   1,
	}

	if browser != "" {
		updateQuery["stats.browser."+browser] = 1
	}

	if os != "" {
		updateQuery["stats.os."+os] = 1
	}

	res := u.DB.Collection("urls").FindOneAndUpdate(ctx,
		bson.M{"short": short},
		bson.M{"$inc": updateQuery},
	)

	return res.Err()
}

//GetURLSBySession get urls from a session
func (u *URLRepository) GetURLSBySession(session string, limit int) ([]models.Url, error) {
	cursor, err := u.DB.Collection("urls").Aggregate(context.TODO(), bson.A{
		bson.M{
			"$match": bson.M{
				"session": bson.M{
					"$eq": session,
				},
			},
		},
		bson.M{"$limit": limit},
		bson.M{"$sort": bson.M{"created_at": -1}},
	})

	if err != nil {
		return nil, err
	}

	var urls []models.Url

	err = cursor.All(context.Background(), &urls)

	return urls, err
}

//GetUrlByCode get url info
func (u *URLRepository) GetUrlByCode(code string) (*models.Url, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	urlDoc := u.DB.Collection("urls").FindOne(ctx, bson.M{"code": code})

	if urlDoc.Err() != nil {
		return nil, urlDoc.Err()
	}

	url := models.Url{}

	err := urlDoc.Decode(&url)

	if err != nil {
		return nil, err
	}

	return &url, nil
}
