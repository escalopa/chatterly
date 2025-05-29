package db

import (
	"context"
	"errors"
	"time"

	"github.com/escalopa/chatterly/internal/domain"
	"github.com/escalopa/chatterly/internal/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	appName = "chatterly"
)

type DB struct {
	users *mongo.Collection

	close func(ctx context.Context) error
}

func New(ctx context.Context, uri string) (*DB, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, errors.New("connect to mongodb: " + err.Error())
	}

	database := client.Database(appName)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var res bson.M
	err = database.RunCommand(ctx, bson.D{{"ping", 1}}).Decode(&res)
	if err != nil {
		return nil, errors.New("ping mongodb: " + err.Error())
	}

	users := database.Collection("users")

	db := &DB{
		users: users,
		close: func(ctx context.Context) error {
			return client.Disconnect(ctx)
		},
	}

	return db, nil
}

func (db *DB) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	f := bson.M{"_id": userID}
	user := &domain.User{}

	err := db.users.FindOne(ctx, f).Decode(user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrDBUserNotFound
		}
		log.Error("db.GetUser", log.Err(err))
		return nil, domain.ErrDBQuery
	}

	return user, nil
}

func (db *DB) CreateUser(ctx context.Context, user *domain.User, provider string) (string, error) {
	f := bson.M{
		"email":    user.Email,
		"provider": provider,
	}

	update := bson.M{"$set": bson.M{
		"name":   user.Name,
		"avatar": user.Avatar,
	}}

	var res domain.User

	opts := options.FindOneAndUpdate().SetUpsert(true)
	err := db.users.FindOneAndUpdate(ctx, f, update, opts).Decode(&res)
	if err != nil {
		log.Error("db.CreateUser", log.Err(err))
		return "", domain.ErrDBQuery
	}

	return res.ID, nil
}

func (db *DB) SetUsername(ctx context.Context, userID, username string) error {
	f := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"username": username}}

	res, err := db.users.UpdateOne(ctx, f, update)
	if err != nil {
		log.Error("db.SetUsername", log.Err(err))
		return domain.ErrDBQuery
	}

	if res.MatchedCount == 0 {
		return domain.ErrDBUserNotFound
	}

	return nil
}

func (db *DB) Close(ctx context.Context) {
	if err := db.close(ctx); err != nil {
		log.Error("db.Close", log.Err(err))
		return
	}
	log.Warn("database connection closed")
}
