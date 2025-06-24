package pagitmongo

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Adapter implements pagination for MongoDB
type Adapter[T any] struct {
	collection *mongo.Collection
	filter     bson.M
	sort       bson.D
}

// NewAdapter creates a new MongoDB pagination adapter
func NewAdapter[T any](collection *mongo.Collection) *Adapter[T] {
	return &Adapter[T]{
		collection: collection,
		filter:     bson.M{},
		sort:       bson.D{{Key: "_id", Value: -1}}, // Default sort by _id descending
	}
}

// WithFilter adds a filter to the query
func (a *Adapter[T]) WithFilter(filter bson.M) *Adapter[T] {
	a.filter = filter
	return a
}

// WithSort sets the sort order
func (a *Adapter[T]) WithSort(sort bson.D) *Adapter[T] {
	a.sort = sort
	return a
}

// Count returns the total number of documents
func (a *Adapter[T]) Count(ctx context.Context) (int64, error) {
	return a.collection.CountDocuments(ctx, a.filter)
}

// Fetch retrieves items for offset-based pagination
func (a *Adapter[T]) Fetch(ctx context.Context, offset, limit int) ([]T, error) {
	opts := options.Find().
		SetSort(a.sort).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := a.collection.Find(ctx, a.filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []T
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return items, nil
}

// FetchWithCursor retrieves items for cursor-based pagination
func (a *Adapter[T]) FetchWithCursor(ctx context.Context, cursor string, limit int) ([]T, string, error) {
	filter := a.filter

	// Decode cursor if provided
	if cursor != "" {
		cursorData, err := decodeCursor(cursor)
		if err != nil {
			return nil, "", err
		}

		// Add cursor condition to filter
		filter = bson.M{
			"$and": []bson.M{
				a.filter,
				cursorData,
			},
		}
	}

	opts := options.Find().
		SetSort(a.sort).
		SetLimit(int64(limit))

	cur, err := a.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, "", err
	}
	defer cur.Close(ctx)

	var items []T
	if err := cur.All(ctx, &items); err != nil {
		return nil, "", err
	}

	// Generate next cursor from last item
	var nextCursor string
	if len(items) == limit {
		nextCursor, err = encodeCursor(items[len(items)-1], a.sort)
		if err != nil {
			return nil, "", err
		}
	}

	return items, nextCursor, nil
}

// CursorData represents the cursor information
type CursorData struct {
	ID   primitive.ObjectID     `bson:"_id,omitempty"`
	Sort map[string]interface{} `bson:"sort,omitempty"`
}

// encodeCursor creates a cursor string from the last item
func encodeCursor(item interface{}, sort bson.D) (string, error) {
	// Extract cursor fields from item based on sort
	// This is a simplified implementation
	data, err := bson.Marshal(item)
	if err != nil {
		return "", err
	}

	var doc bson.M
	if err := bson.Unmarshal(data, &doc); err != nil {
		return "", err
	}

	cursorData := make(map[string]interface{})
	for _, sortField := range sort {
		if val, ok := doc[sortField.Key]; ok {
			cursorData[sortField.Key] = val
		}
	}

	jsonData, err := json.Marshal(cursorData)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(jsonData), nil
}

// decodeCursor decodes a cursor string into filter conditions
func decodeCursor(cursor string) (bson.M, error) {
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}

	var cursorData map[string]interface{}
	if err := json.Unmarshal(data, &cursorData); err != nil {
		return nil, err
	}

	// Build filter based on cursor data
	// This is a simplified implementation for _id-based cursors
	if id, ok := cursorData["_id"]; ok {
		return bson.M{"_id": bson.M{"$gt": id}}, nil
	}

	return bson.M{}, nil
}

// AggregationAdapter implements pagination using MongoDB aggregation pipeline
type AggregationAdapter[T any] struct {
	collection *mongo.Collection
	pipeline   mongo.Pipeline
	countStage bson.D
}

// NewAggregationAdapter creates a new MongoDB aggregation pagination adapter
func NewAggregationAdapter[T any](collection *mongo.Collection, pipeline mongo.Pipeline) *AggregationAdapter[T] {
	return &AggregationAdapter[T]{
		collection: collection,
		pipeline:   pipeline,
		countStage: bson.D{{Key: "$count", Value: "total"}},
	}
}

// Count returns the total number of documents in the pipeline
func (a *AggregationAdapter[T]) Count(ctx context.Context) (int64, error) {
	// Create count pipeline
	countPipeline := append(a.pipeline, a.countStage)

	cursor, err := a.collection.Aggregate(ctx, countPipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Total int64 `bson:"total"`
	}
	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) > 0 {
		return results[0].Total, nil
	}
	return 0, nil
}

// Fetch retrieves items for offset-based pagination
func (a *AggregationAdapter[T]) Fetch(ctx context.Context, offset, limit int) ([]T, error) {
	// Add pagination stages to pipeline
	paginatedPipeline := append(a.pipeline,
		bson.D{{Key: "$skip", Value: offset}},
		bson.D{{Key: "$limit", Value: limit}},
	)

	cursor, err := a.collection.Aggregate(ctx, paginatedPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []T
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return items, nil
}

// FetchWithCursor is not easily implemented for aggregation pipelines
func (a *AggregationAdapter[T]) FetchWithCursor(ctx context.Context, cursor string, limit int) ([]T, string, error) {
	// This would require modifying the pipeline to add cursor-based filtering
	// Implementation depends on specific use case
	return nil, "", nil
}
