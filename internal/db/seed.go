package db

import (
	"context"
	"log"

	"github.com/MohummedSoliman/social/internal/store"
)

var usernames = []string{
	"brave_tiger", "cool_dragon", "silent_wolf", "fast_eagle", "wild_fox", "smart_lion", "bright_bear",
}

var titles = []string{
	"admin", "editor", "viewer", "guest", "author", "manager", "owner",
}

var contents = []string{
	"daily report", "project update", "team meeting", "new feature", "bug fix", "code review", "design mockup",
}

var tags = []string{
	"golang", "backend", "database", "api", "cloud", "docker", "performance",
}

var comments = []string{
	"Great work on this!",
	"I think we should review this part.",
	"Looks good to me.",
	"Can we optimize this section?",
	"Please double-check the logic here.",
	"Nice improvement over the last version.",
	"Let’s discuss this in the next meeting.",
}

const numberOfSeedData = 7

func Seed(store store.Storage) {
	ctx := context.Background()

	users := generateUsers()
	for _, user := range users {
		if err := store.Users.Create(ctx, user); err != nil {
			log.Println("Error creating user", err)
			return
		}
	}

	posts := generatePosts(users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("Error creating post", err)
			return
		}
	}

	comments := generateComments(posts)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Println("Error creating post", err)
			return
		}
	}

	log.Println("Seeding has been completed!!!!")
}

func generateComments(posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, numberOfSeedData)

	for i := range comments {
		comments[i] = &store.Comment{
			UserID:  posts[i].UserID,
			PostID:  posts[i].ID,
			Content: contents[i],
		}
	}
	return comments
}

func generatePosts(users []*store.User) []*store.Post {
	posts := make([]*store.Post, numberOfSeedData)
	for i := range posts {
		user := users[i]
		posts[i] = &store.Post{
			UserID:  user.ID,
			Title:   titles[i],
			Content: contents[i],
			Tags:    []string{tags[i]},
		}
	}
	return posts
}

func generateUsers() []*store.User {
	users := make([]*store.User, numberOfSeedData)

	for i := range users {
		users[i] = &store.User{
			Username: usernames[i],
			Email:    usernames[i] + "@example.com",
			Password: "123123",
		}
	}
	return users
}
