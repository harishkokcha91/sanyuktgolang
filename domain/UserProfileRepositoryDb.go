package domain

import (
	"context"
	"fmt"
	"sanyuktgolang/errs"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserProfileRepositoryDb struct {
	client *mongo.Client
}

func (d UserProfileRepositoryDb) FindAll(status string) ([]UserProfile, *errs.AppError) {
	collection := d.client.Database("mydb").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var users []UserProfile

	results, err := collection.Find(ctx, bson.M{})

	if err != nil {
		print("errrt1")
	}

	//reading from the db in an optimal way
	defer results.Close(ctx)
	for results.Next(ctx) {
		fmt.Println(results)
		var singleUser UserProfile
		if err = results.Decode(&singleUser); err != nil {
			print("errrt")
		}

		users = append(users, singleUser)
	}
	fmt.Println(users)
	defer cancel()

	return nil, nil
}

func NewUserProfileRepositoryDb(dbClient *mongo.Client) UserProfileRepositoryDb {
	return UserProfileRepositoryDb{dbClient}
}
