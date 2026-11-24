package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

const (
	maxOutboxAttempts  = 10
	outboxBatchSize    = 20
	outboxLeaseSeconds = 60
	outboxRetentionDay = 90
)

const flashSaleStartedSQL = `
	INSERT INTO outbox_events (event_type, aggregate_id, payload)
	SELECT 'flash_sale.started', CAST(fs.id AS CHAR),
	       JSON_OBJECT('flash_sale_id', fs.id, 'name', fs.name,
	                   'start_time', DATE_FORMAT(fs.start_time, '%Y-%m-%dT%H:%i:%sZ'),
	                   'end_time',   DATE_FORMAT(fs.end_time,   '%Y-%m-%dT%H:%i:%sZ'))
	FROM flash_sales fs
	WHERE fs.status NOT IN ('cancelled', 'ended')
	  AND NOW() >= fs.start_time AND NOW() < fs.end_time
	ON DUPLICATE KEY UPDATE outbox_events.id = outbox_events.id`

const flashSaleEndedSQL = `
	INSERT INTO outbox_events (event_type, aggregate_id, payload)
	SELECT 'flash_sale.ended', CAST(fs.id AS CHAR),
	       JSON_OBJECT('flash_sale_id', fs.id, 'name', fs.name,
	                   'end_time', DATE_FORMAT(fs.end_time, '%Y-%m-%dT%H:%i:%sZ'))
	FROM flash_sales fs
	WHERE fs.status <> 'cancelled'
	  AND NOW() >= fs.end_time
	  AND fs.end_time >= DATE_SUB(NOW(), INTERVAL 1 DAY)
	ON DUPLICATE KEY UPDATE outbox_events.id = outbox_events.id`

type OutboxService struct {
	db     *sql.DB
	logger *zap.Logger
	client *http.Client
	url    string
	secret string
}

func InitOutboxService(db *sql.DB, logger *zap.Logger) *OutboxService {
	s := &OutboxService{
		db:     db,
		logger: logger,
		client: &http.Client{Timeout: 5 * time.Second},
		url:    os.Getenv("N8N_WEBHOOK_URL"),
		secret: os.Getenv("N8N_WEBHOOK_SECRET"),
	}
	if s.url == "" {
		logger.Warn("N8N_WEBHOOK_URL is empty: outbox events are queued but not delivered")
	}
	return s
}

func (s *OutboxService) Run(ctx context.Context) {
	enqueue := time.NewTicker(30 * time.Second)
	dispatch := time.NewTicker(5 * time.Second)
	cleanup := time.NewTicker(time.Hour)
	defer enqueue.Stop()
	defer dispatch.Stop()
	defer cleanup.Stop()

	fmt.Println("outbox event started")

	for {
		select {
		case <-ctx.Done():
			return
		case <-enqueue.C:
			fmt.Println("enqueue called")
			s.EnqueueFlashSaleEvents(ctx)
		case <-dispatch.C:
			if s.url != "" {
				fmt.Println("dispatch called")
				s.DispatchPending(ctx)
			}
		case <-cleanup.C:
			fmt.Println("cleanup called")
			s.Cleanup(ctx)
		}
	}
}

func (s *OutboxService) EnqueueFlashSaleEvents(ctx context.Context) {
	for _, q := range []string{flashSaleStartedSQL, flashSaleEndedSQL} {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			s.logger.Error("enqueue flash sale events", zap.Error(err))
		}
	}
}

type outboxRow struct {
	ID          int64
	Type        string
	AggregateID string
	Payload     []byte
	Attempts    int
	CreatedAt   time.Time
}

func (s *OutboxService) DispatchPending(ctx context.Context) {
	var dbName, tz string
	var now time.Time
	var total, due int
	_ = s.db.QueryRowContext(ctx, `SELECT DATABASE(), NOW(), @@session.time_zone`).Scan(&dbName, &now, &tz)
	_ = s.db.QueryRowContext(ctx, `
	SELECT COUNT(*),
	       COALESCE(SUM(sent_at IS NULL AND attempts < 10 AND next_attempt_at <= NOW()), 0)
	FROM outbox_events`).Scan(&total, &due)
	fmt.Println("db:", dbName, "| now:", now, "| tz:", tz, "| total:", total, "| due:", due)

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, event_type, aggregate_id, payload, attempts, created_at
		FROM outbox_events
		WHERE sent_at IS NULL AND attempts < ? AND next_attempt_at <= NOW()
		ORDER BY id LIMIT ?`, maxOutboxAttempts, outboxBatchSize)
	if err != nil {
		s.logger.Error("query outbox", zap.Error(err))
		return
	}
	var batch []outboxRow

	for rows.Next() {
		var e outboxRow
		if err := rows.Scan(&e.ID, &e.Type, &e.AggregateID, &e.Payload, &e.Attempts, &e.CreatedAt); err != nil {
			s.logger.Error("scan outbox", zap.Error(err))
			continue
		}
		batch = append(batch, e)
	}
	rows.Close() // tutup sebelum menjalankan query lain
	if err := rows.Err(); err != nil {
		s.logger.Error("iterate outbox", zap.Error(err))
		return
	}

	// Setelah pengiriman, tulis hasilnya walau ctx sudah dibatalkan, supaya event tidak terkirim dua kali saat shutdown.
	done := context.WithoutCancel(ctx)

	for _, e := range batch {
		if ctx.Err() != nil {
			return
		}

		// Klaim atomik (lease): hanya satu instance yang mendapat RowsAffected = 1.
		res, err := s.db.ExecContext(ctx, `
			UPDATE outbox_events
			SET attempts = attempts + 1,
			    next_attempt_at = DATE_ADD(NOW(), INTERVAL ? SECOND)
			WHERE id = ? AND sent_at IS NULL AND next_attempt_at <= NOW()`,
			outboxLeaseSeconds, e.ID)
		if err != nil {
			s.logger.Error("claim outbox", zap.Int64("id", e.ID), zap.Error(err))
			continue
		}
		if n, _ := res.RowsAffected(); n != 1 {
			continue
		}

		if err := s.send(ctx, e); err != nil {
			delay := outboxBackoff(e.Attempts + 1)
			s.logger.Warn("deliver outbox failed", zap.Int64("id", e.ID), zap.String("type", e.Type),
				zap.Int("attempt", e.Attempts+1), zap.Duration("retry_in", delay), zap.Error(err))
			_, _ = s.db.ExecContext(done, `
				UPDATE outbox_events
				SET next_attempt_at = DATE_ADD(NOW(), INTERVAL ? SECOND), last_error = ?
				WHERE id = ?`, int(delay.Seconds()), truncate(err.Error(), 250), e.ID)
			continue
		}

		if _, err := s.db.ExecContext(done,
			`UPDATE outbox_events SET sent_at = NOW(), last_error = NULL WHERE id = ?`, e.ID); err != nil {
			s.logger.Error("mark outbox sent", zap.Int64("id", e.ID), zap.Error(err))
		}
	}
}

func (s *OutboxService) send(ctx context.Context, e outboxRow) error {
	body, err := json.Marshal(map[string]any{
		"id":           e.ID,
		"type":         e.Type,
		"aggregate_id": e.AggregateID,
		"data":         json.RawMessage(e.Payload),
		"created_at":   e.CreatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}

	fmt.Println("Body send to webhook: ", body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.secret)
	req.Header.Set("Idempotency-Key", e.Type+":"+e.AggregateID) // untuk dedupe di n8n

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// 10s, 20s, 40s, ... maksimal 15 menit
func outboxBackoff(attempt int) time.Duration {
	d := time.Duration(int64(1)<<min(attempt, 8)) * 5 * time.Second
	return min(d, 15*time.Minute)
}

func (s *OutboxService) Cleanup(ctx context.Context) {
	if _, err := s.db.ExecContext(ctx, `
		DELETE FROM outbox_events
		WHERE sent_at IS NOT NULL AND sent_at < DATE_SUB(NOW(), INTERVAL ? DAY)
		LIMIT 1000`, outboxRetentionDay); err != nil {
		s.logger.Error("cleanup outbox", zap.Error(err))
	}
}
