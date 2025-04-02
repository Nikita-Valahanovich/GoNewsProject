package storage

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	_, err := New()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDB_News(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	posts := []Post{
		{
			Title: "Test Post",
			Link:  strconv.Itoa(rand.Intn(1_000_000_000)),
		},
	}
	db, err := New()
	if err != nil {
		t.Fatal(err)
	}
	err = db.StoreNews(posts)
	if err != nil {
		t.Fatal(err)
	}
	news, err := db.News(2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", news)
}
