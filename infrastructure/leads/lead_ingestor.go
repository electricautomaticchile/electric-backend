package leads

import (
	"context"
	"electric-backend/config"
	"electric-backend/infrastructure/entities"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LeadIngestor struct {
	collection    *mongo.Collection
	queue         chan entities.LeadEntity
	batchSize     int
	workers       int
	flushInterval time.Duration
	writeTimeout  time.Duration
	maxRetries    int
	stop          chan struct{}
	stopOnce      sync.Once
	wg            sync.WaitGroup
	enqueued      atomic.Uint64
	flushed       atomic.Uint64
	dropped       atomic.Uint64
	failed        atomic.Uint64
}

var (
	defaultMu       sync.RWMutex
	defaultIngestor *LeadIngestor
)

func StartDefaultLeadIngestor(db *mongo.Database, cfg *config.Config) {
	if db == nil || cfg == nil || !cfg.LeadIngestAsync {
		return
	}

	ingestor := NewLeadIngestor(
		db.Collection(entities.LeadEntity{}.CollectionName()),
		cfg.LeadIngestQueueSize,
		cfg.LeadIngestBatchSize,
		cfg.LeadIngestWorkers,
		cfg.LeadIngestFlushInterval,
		cfg.LeadIngestWriteTimeout,
		cfg.LeadIngestMaxRetries,
	)
	ingestor.Start()

	defaultMu.Lock()
	defaultIngestor = ingestor
	defaultMu.Unlock()

	log.Printf(
		"✅ Lead async ingestor activo (queue=%d batch=%d workers=%d flush=%s write_timeout=%s retries=%d)",
		cfg.LeadIngestQueueSize,
		cfg.LeadIngestBatchSize,
		cfg.LeadIngestWorkers,
		cfg.LeadIngestFlushInterval,
		cfg.LeadIngestWriteTimeout,
		cfg.LeadIngestMaxRetries,
	)
}

func StopDefaultLeadIngestor(ctx context.Context) {
	defaultMu.RLock()
	ingestor := defaultIngestor
	defaultMu.RUnlock()
	if ingestor == nil {
		return
	}
	ingestor.Stop(ctx)
}

func DefaultLeadIngestor() *LeadIngestor {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultIngestor
}

func NewLeadIngestor(collection *mongo.Collection, queueSize int, batchSize int, workers int, flushInterval time.Duration, writeTimeout time.Duration, maxRetries int) *LeadIngestor {
	if queueSize < 1 {
		queueSize = 50000
	}
	if batchSize < 1 {
		batchSize = 500
	}
	if workers < 1 {
		workers = 1
	}
	if flushInterval <= 0 {
		flushInterval = 100 * time.Millisecond
	}
	if writeTimeout <= 0 {
		writeTimeout = 15 * time.Second
	}
	if maxRetries < 1 {
		maxRetries = 1
	}

	return &LeadIngestor{
		collection:    collection,
		queue:         make(chan entities.LeadEntity, queueSize),
		batchSize:     batchSize,
		workers:       workers,
		flushInterval: flushInterval,
		writeTimeout:  writeTimeout,
		maxRetries:    maxRetries,
		stop:          make(chan struct{}),
	}
}

func (i *LeadIngestor) Start() {
	for worker := 0; worker < i.workers; worker++ {
		i.wg.Add(1)
		go i.runWorker()
	}
}

func (i *LeadIngestor) Stop(ctx context.Context) {
	i.stopOnce.Do(func() {
		close(i.stop)
	})

	done := make(chan struct{})
	go func() {
		i.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		log.Printf("⚠️ Lead ingestor no alcanzo a vaciar cola: %v", ctx.Err())
	}
}

func (i *LeadIngestor) Enqueue(lead entities.LeadEntity) bool {
	if lead.ID.IsZero() || lead.Email == "" || lead.Name == "" {
		i.dropped.Add(1)
		return false
	}

	select {
	case i.queue <- lead:
		i.enqueued.Add(1)
		return true
	default:
		i.dropped.Add(1)
		return false
	}
}

func (i *LeadIngestor) Stats() map[string]uint64 {
	return map[string]uint64{
		"enqueued": i.enqueued.Load(),
		"flushed":  i.flushed.Load(),
		"dropped":  i.dropped.Load(),
		"failed":   i.failed.Load(),
		"queued":   uint64(len(i.queue)),
	}
}

func (i *LeadIngestor) runWorker() {
	defer i.wg.Done()

	ticker := time.NewTicker(i.flushInterval)
	defer ticker.Stop()

	pending := make([]interface{}, 0, i.batchSize)
	for {
		select {
		case lead := <-i.queue:
			pending = append(pending, lead)
			if len(pending) >= i.batchSize {
				i.flush(pending)
				pending = make([]interface{}, 0, i.batchSize)
			}
		case <-ticker.C:
			if len(pending) > 0 {
				i.flush(pending)
				pending = make([]interface{}, 0, i.batchSize)
			}
		case <-i.stop:
			pending = i.drainInto(pending)
			if len(pending) > 0 {
				i.flush(pending)
			}
			return
		}
	}
}

func (i *LeadIngestor) drainInto(pending []interface{}) []interface{} {
	for {
		select {
		case lead := <-i.queue:
			pending = append(pending, lead)
			if len(pending) >= i.batchSize {
				i.flush(pending)
				pending = make([]interface{}, 0, i.batchSize)
			}
		default:
			return pending
		}
	}
}

func (i *LeadIngestor) flush(pending []interface{}) {
	if len(pending) == 0 {
		return
	}

	var lastErr error
	for attempt := 1; attempt <= i.maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), i.writeTimeout)
		_, err := i.collection.InsertMany(ctx, pending, options.InsertMany().SetOrdered(false))
		cancel()
		if err == nil || duplicateKeyOnly(err) {
			i.flushed.Add(uint64(len(pending)))
			return
		}
		lastErr = err
		if attempt < i.maxRetries {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}
	}

	i.failed.Add(uint64(len(pending)))
	log.Printf("⚠️ error escribiendo batch Leads tras %d intento(s) (%d leads): %v", i.maxRetries, len(pending), lastErr)
}

func duplicateKeyOnly(err error) bool {
	var bulkErr mongo.BulkWriteException
	if !errors.As(err, &bulkErr) || len(bulkErr.WriteErrors) == 0 {
		return false
	}
	for _, writeErr := range bulkErr.WriteErrors {
		if writeErr.Code != 11000 {
			return false
		}
	}
	return true
}
