package main

import (
	"ads-system/internal/api"
	"ads-system/internal/database"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rodrwan/secretly/pkg/secretly"
)

func main() {
	envs := secretly.New(
		secretly.WithBaseURL("http://environment:9000"),
	)
	if err := envs.LoadToEnvironment("ad-server"); err != nil {
		log.Fatalf("Error al cargar el archivo .env: %v", err)
	}

	srv := gin.Default()

	// add json logger to gin
	srv.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method, param.Path, param.Request.Proto,
			param.StatusCode, param.Latency, param.Request.UserAgent(), param.ErrorMessage,
		)
	}))

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

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	var redisClient = redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	// Crear el cliente de base de datos
	db := database.New(pool)
	// Crear el servidor de la API
	api := api.NewServer(db, redisClient)

	api.Routes(srv)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv.Run(":" + port)
}
