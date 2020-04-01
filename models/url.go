package models

import (
	"bytes"
	"math/rand"
	"time"
)

type Url struct {
	Original    string `form:"url" bson:"original_url" binding:"required"` // Url to be shortened
	Session     string `bson:"session"`
	DeleteAfter *int   `form:"delete_after" bson:"delete_after" binding:"required,min=0"` // Delete after X minutes. if it's 0, then it will not be automatically deleted

	Short     string    `bson:"short_url"`  // Shortened url
	Code      string    `bson:"code"`       // Used to delete and get url info
	CreatedAt time.Time `bson:"created_at"` //When this url was created
}

func (url *Url) Shorten() {
	url.Short = randomString(6)
	url.Code = randomString(6)
}

// Expired returns if the url have expired and should be deleted
func (url Url) Expired() bool {
	if url.DeleteAfter == nil {
		return false
	}

	return time.Now().After(url.CreatedAt.Add(time.Minute * time.Duration(*url.DeleteAfter)))
}

/// https://github.com/thanhpk/randstr
// list of default letters that can be used to make a random string when calling String
// function with no letters provided
var defLetters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// String generates a random string using only letters provided in the letters parameter
// if user ommit letters parameters, this function will use defLetters instead
func randomString(n int, letters ...string) string {
	var letterRunes []rune
	if len(letters) == 0 {
		letterRunes = defLetters
	} else {
		letterRunes = []rune(letters[0])
	}

	var bb bytes.Buffer
	bb.Grow(n)

	// on each loop, generate one random rune and append to output
	for i := 0; i < n; i++ {
		bb.WriteRune(letterRunes[rand.Intn(len(letterRunes))])
	}
	return bb.String()
}
