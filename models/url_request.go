package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type UrlRequest struct {
	UrlShort    string    `bson:"url_short"`
	IP          bson.M    `bson:"ip"`
	Referer     string    `bson:"referer"`
	RequestTime time.Time `bson:"time"`
	UserAgent   bson.M    `bson:"user_agent"`
}

func NewURLRequest(urlShort, referer string, ipData, userAgent bson.M) UrlRequest {
	return UrlRequest{
		urlShort,
		bson.M{
			"country":     ipData["country"],
			"countryCode": ipData["countryCode"],
			"city":        ipData["city"],
			"region":      ipData["region"],
			"ip":          ipData["query"],
		},
		referer,
		time.Now(),
		userAgent,
	}
}
