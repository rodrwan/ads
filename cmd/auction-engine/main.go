package main

import (
	"ads-system/internal/database"
	"context"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/rodrwan/secretly/pkg/secretly"
)

func main() {
	envs := secretly.New(
		secretly.WithBaseURL("http://environment:9000"),
	)
	if err := envs.LoadToEnvironment("auction-engine"); err != nil {
		log.Fatalf("Error al cargar el archivo .env: %v", err)
	}

	// Configurar la conexión a la base de datos
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/themenu?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}
	defer pool.Close()

	// Crear el cliente de base de datos
	db := database.New(pool)

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	go ProcessImpressions(db, redisClient)
	go ProcessClicks(db, redisClient)

	select {} // keep running
}

func ProcessImpressions(db database.Querier, redisClient *redis.Client) {
	group := "impression-group"
	consumer := "worker-1"
	stream := "stream:impressions"
	ctx := context.Background()

	redisClient.XGroupCreateMkStream(ctx, stream, group, "0")

	for {
		msgs, err := redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Count:    10,
			Block:    0,
		}).Result()

		if err != nil {
			log.Println("Error reading impressions:", err)
			continue
		}

		for _, msg := range msgs[0].Messages {
			adID := msg.Values["ad_id"].(pgtype.UUID)
			placementID := msg.Values["placement_id"].(pgtype.UUID)
			auctionID := msg.Values["auction_id"].(pgtype.UUID)
			userContext := msg.Values["user_context"].([]byte)

			var id pgtype.UUID
			_ = id.Scan(uuid.New().String())

			_, err = db.CreateImpression(ctx, database.CreateImpressionParams{
				ID:          id,
				AdID:        adID,
				PlacementID: placementID,
				AuctionID:   auctionID,
				UserContext: userContext,
			})

			if err != nil {
				log.Println("DB insert error (impression):", err)
				continue
			}

			redisClient.XAck(ctx, stream, group, msg.ID)
		}
	}
}

func ProcessClicks(db database.Querier, redisClient *redis.Client) {
	group := "click-group"
	consumer := "worker-1"
	stream := "stream:clicks"
	ctx := context.Background()

	redisClient.XGroupCreateMkStream(ctx, stream, group, "0")

	for {
		msgs, err := redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Count:    10,
			Block:    0,
		}).Result()

		if err != nil {
			log.Println("Error reading clicks:", err)
			continue
		}

		for _, msg := range msgs[0].Messages {
			impressionID := msg.Values["impression_id"].(pgtype.UUID)

			var id pgtype.UUID
			_ = id.Scan(uuid.New().String())

			_, err := db.CreateClick(ctx, database.CreateClickParams{
				ID:           id,
				ImpressionID: impressionID,
			})

			if err != nil {
				log.Println("DB insert error (click):", err)
				continue
			}

			redisClient.XAck(ctx, stream, group, msg.ID)
		}
	}
}
