package repository

import (
	"context"
	"time"

	"github.com/arthurdm10/url-shortener/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type URLRequestRepository struct {
	DB *mongo.Database
}

// CreateURLRequest creates a new urlRequest in the database
func (r *URLRequestRepository) CreateURLRequest(ur models.UrlRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := r.DB.Collection("requests").InsertOne(ctx, ur)
	return err
}

// DeleteByURL deletes all requests to a URL
func (r *URLRequestRepository) DeleteByURL(short string) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_, err := r.DB.Collection("requests").DeleteMany(ctx, bson.M{"url_short": short})
	return err
}

// GetRequestsByURL get all requests made to a URL
func (r *URLRequestRepository) GetRequestsByURL(short string) ([]models.UrlRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	res, err := r.DB.Collection("requests").Find(ctx, bson.M{"short_url": short})
	if err != nil {
		return nil, err
	}

	var requests []models.UrlRequest

	err = res.All(context.Background(), &requests)

	return requests, err
}
