package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/theabgarg/market-aggregator/internal/domain"
)

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(ctx context.Context, dbUrl string) (*PostgresRepo, error) {
	pool, err := pgxpool.New(ctx, dbUrl)
	if err !=nil {
		return nil, fmt.Errorf("Error creating new pool %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error while pinging the database %w", err)
	}

	createTableSQL := `
		CREATE TABLE IF NOT EXISTS ticks (
			id SERIAL PRIMARY KEY,
			symbol VARCHAR(20) NOT NULL,
			price DOUBLE PRECISION NOT NULL,
			volume DOUBLE PRECISION NOT NULL,
			source VARCHAR(50) NOT NULL,
			timestamp TIMESTAMPTZ NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_symbol_time ON ticks (symbol, timestamp DESC);
	`

	if _, err := pool.Exec(ctx, createTableSQL); err != nil {
		return nil, fmt.Errorf("Failed to create table %w", err)
	}

	slog.Info("connected to databse and verified schema")
	return &PostgresRepo{pool : pool}, nil
}

func (r *PostgresRepo) Close(){
	r.pool.Close()
}

func (r *PostgresRepo) InsertTick(ctx context.Context, tick domain.MarketData) error {
	query := `INSERT INTO ticks (symbol, price, volume, source, timestamp) 
	VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, tick.Symbol, tick.Price, tick.Volume, tick.Source, tick.Timestamp)
	return err
}

func (r *PostgresRepo) GetHistoricalTicks(ctx context.Context, symbol string, limit int) ([]domain.MarketData, error) {
	query := `
		SELECT symbol, price, volume, source, timestamp
		FROM ticks
		WHERE symbol = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	ticks := make([]domain.MarketData, 0, limit)

	for rows.Next() {
		var tick domain.MarketData

		err := rows.Scan(&tick.Symbol, &tick.Price, &tick.Volume, &tick.Source, &tick.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("Failed to scan row: %w", err)
		}
		ticks = append(ticks, tick)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return ticks, nil

}