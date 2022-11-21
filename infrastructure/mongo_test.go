package infrastructure

import (
	"context"
	"fmt"
	"testing"

	"github.com/sergicanet9/scv-go-tools/v3/mocks"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

const testEntityName = "test"

type testEntity struct {
	ID string `bson:"_id,omitempty"`
}

// TestConnectMongoDB_InvalidConnection checks that ConnectMongoDB returns an error when an invalid connection string is provided
func TestConnectMongoDB_InvalidConnection(t *testing.T) {
	// Arrange
	expectedError := "an unexpected error happened while opening the connection: error parsing uri: scheme must be \"mongodb\" or \"mongodb+srv\""

	// Act
	_, err := ConnectMongoDB(context.Background(), "test", "invalid-connection")

	// Assert
	assert.Equal(t, expectedError, err.Error())
}

// TestPingMongo_Ok checks that pingMongo does not return an error when a valid db is received
func TestPingMongo_Ok(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	mt.Run("", func(mt *mtest.T) {
		// Arrange
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		// Act
		_, err := pingMongo(mt.Client, "test", nil)

		// Assert
		assert.Equal(mt, nil, err)
	})
}

// TestPingMongo_NilDB checks that pingMongo returns an error when a nil db is received
func TestPingMongo_NilDB(t *testing.T) {
	// Arrange
	expectedError := "an unexpected error happened while opening the connection: %!s(<nil>)"
	// Act
	_, err := pingMongo(nil, "", nil)

	// Assert
	assert.Equal(t, expectedError, err.Error())
}

// TestCreate_OK checks that Create does not return an error when a valid entity is received
func TestCreate_OK(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		newEntity := testEntity{}

		// Act
		id, err := repo.Create(context.Background(), newEntity)

		// Assert
		assert.IsType(t, newEntity.ID, id)
		assert.Equal(t, nil, err)
	})
}

// TestCreate_InsertOneError checks that Create returns an error when InsertOne fails
func TestCreate_InsertOneError(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})
		newEntity := testEntity{}

		// Act
		_, err := repo.Create(context.Background(), newEntity)

		// Assert
		assert.NotEmpty(t, err)
	})
}

// TestGet_Ok checks that Get does not return an error when a valid filter is received
func TestGet_Ok(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		skip := 1
		take := 1
		get := mtest.CreateCursorResponse(1,
			fmt.Sprintf("test.%s", testEntityName),
			mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: primitive.NewObjectID()}})
		killCursors := mtest.CreateCursorResponse(0, fmt.Sprintf("test.%s", testEntityName), mtest.NextBatch)

		mt.AddMockResponses(get, killCursors)

		// Act
		result, err := repo.Get(context.Background(), map[string]interface{}{}, &skip, &take)

		// Assert
		assert.Equal(t, nil, err)
		assert.True(t, len(result) == 1)

		entity := *(result[0].(*testEntity))
		assert.IsType(t, testEntity{}, entity)
	})
}

// TestGet_FindError checks that Get returns an error when Find fails
func TestGet_FindError(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})

		// Act
		_, err := repo.Get(context.Background(), map[string]interface{}{}, nil, nil)

		// Assert
		assert.NotEmpty(t, err)
	})
}

// TestGet_DecodeEntryError checks that Get returns an error when the result cannot be decoded
func TestGet_DecodeEntryError(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     make(chan int),
		}

		get := mtest.CreateCursorResponse(1,
			fmt.Sprintf("test.%s", testEntityName),
			mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: primitive.NewObjectID()}})
		killCursors := mtest.CreateCursorResponse(0, fmt.Sprintf("test.%s", testEntityName), mtest.NextBatch)

		mt.AddMockResponses(get, killCursors)

		// Act
		_, err := repo.Get(context.Background(), map[string]interface{}{}, nil, nil)

		// Assert
		assert.NotEmpty(t, err)
	})
}

// TestGetByID_Ok checks that GetByID does not return an error when the received ID has a valid format
func TestGetByID_Ok(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		get := mtest.CreateCursorResponse(1,
			fmt.Sprintf("test.%s", testEntityName),
			mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: primitive.NewObjectID()}})
		killCursors := mtest.CreateCursorResponse(0, fmt.Sprintf("test.%s", testEntityName), mtest.NextBatch)

		mt.AddMockResponses(get, killCursors)

		// Act
		result, err := repo.GetByID(context.Background(), primitive.NewObjectID().Hex())

		// Assert
		assert.Equal(t, nil, err)

		entity := *(result.(*testEntity))
		assert.IsType(t, testEntity{}, entity)
	})
}

// TestGetByID_InvalidID checks that GetByID returns an error when the received ID does not have a valid format
func TestGetByID_InvalidID(t *testing.T) {
	mt := mocks.NewMongoDB(t)

	defer mt.Close()
	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     make(chan int),
		}

		// Act
		_, err := repo.GetByID(context.Background(), "invalid-id")

		// Assert
		assert.NotEmpty(t, err)
	})
}

// TestGetByID_FindOneError checks that GetByID returns an error when FindOne fails
func TestGetByID_FindOneError(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})

		// Act
		_, err := repo.GetByID(context.Background(), primitive.NewObjectID().Hex())

		// Assert
		assert.NotEmpty(t, err)
	})
}

// TestUpdate_OK checks that Update does not return an error when the received ID has a valid format
func TestUpdate_OK(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 1},
		})
		newEntity := testEntity{}

		// Act
		err := repo.Update(context.Background(), primitive.NewObjectID().Hex(), newEntity)

		// Assert
		assert.Equal(t, nil, err)
	})
}

// TestUpdate_InvalidID checks that Update returns an error when the received ID does not have a valid format
func TestUpdate_InvalidID(t *testing.T) {
	mt := mocks.NewMongoDB(t)

	defer mt.Close()
	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     make(chan int),
		}
		newEntity := testEntity{}

		// Act
		err := repo.Update(context.Background(), "invalid-id", newEntity)

		// Assert
		assert.NotEmpty(t, err)
	})
}

// TestUpdate_UpdateOneError checks that Update returns an error when UpdateOne fails
func TestUpdate_UpdateOneError(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})
		newEntity := testEntity{}

		// Act
		err := repo.Update(context.Background(), primitive.NewObjectID().Hex(), newEntity)

		// Assert
		assert.NotEmpty(t, err)
	})
}

// TestUpdate_NotUpdatedError checks that Update returns an error when UpdateOne does not update any document
func TestUpdate_NotUpdatedError(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 0},
		})
		newEntity := testEntity{}

		// Act
		err := repo.Update(context.Background(), primitive.NewObjectID().Hex(), newEntity)

		// Assert
		assert.Equal(t, mongo.ErrNoDocuments, err)
	})
}

// TestDelete_OK checks that Delete does not return an error when the received ID has a valid format
func TestDelete_OK(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 1},
		})

		// Act
		err := repo.Delete(context.Background(), primitive.NewObjectID().Hex())

		// Assert
		assert.Equal(t, nil, err)
	})
}

// TestDelete_InvalidID checks that Delete returns an error when the received ID does not have a valid format
func TestDelete_InvalidID(t *testing.T) {
	mt := mocks.NewMongoDB(t)

	defer mt.Close()
	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     make(chan int),
		}

		// Act
		err := repo.Delete(context.Background(), "invalid-id")

		// Assert
		assert.NotEmpty(t, err)
	})
}

// TestDelete_DeleteOneError checks that Delete returns an error when DeleteOne fails
func TestDelete_DeleteOneError(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})

		// Act
		err := repo.Delete(context.Background(), primitive.NewObjectID().Hex())

		// Assert
		assert.NotEmpty(t, err)
	})
}

// TestDelete_NotDeletedError checks that Delete returns an error when DeleteOne does not delete any document
func TestDelete_NotDeletedError(t *testing.T) {
	mt := mocks.NewMongoDB(t)
	defer mt.Close()

	mt.Run("", func(mt *mtest.T) {
		// Arrange
		repo := MongoRepository{
			DB:         mt.DB,
			Collection: mt.DB.Collection(testEntityName),
			Target:     testEntity{},
		}

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 0},
		})

		// Act
		err := repo.Delete(context.Background(), primitive.NewObjectID().Hex())

		// Assert
		assert.Equal(t, mongo.ErrNoDocuments, err)
	})
}