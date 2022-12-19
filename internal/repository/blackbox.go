package repository

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/gtgaleevtimur/gofermart/internal/entity"
)

type Blackbox struct {
	url       string
	storage   entity.Storager
	limit     uint32
	needSleep int32
	pool      map[uint64]*entity.Order
}

type blackboxOrder struct {
	*Blackbox
	ctx   context.Context
	order *entity.Order
}

type blackboxOrderX struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func (bo *blackboxOrder) Do() error {
	ctx, cancel := context.WithTimeout(bo.ctx, 60*time.Second)
	defer cancel()
	order := bo.order
	url := fmt.Sprintf("%s%d", bo.url, order.ID)
	log.Debug().Str("making request", url)

	ao := &blackboxOrderX{}
	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", "*/*").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Length", "0").
		SetContext(ctx).
		SetResult(&ao).
		Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusInternalServerError {
		return fmt.Errorf("internal server error, status code %d", resp.StatusCode())
	}
	if resp.StatusCode() == http.StatusTooManyRequests {
		n := uint32(rand.Intn(10)) + 2
		atomic.StoreUint32(&bo.limit, n)
		log.Warn().Str("too many requests detected", string(resp.Body()))
		return ErrTooManyRequests
	}
	if resp.StatusCode() == http.StatusNoContent {
		log.Warn().Uint64("no content for order", order.ID)
		return nil
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unknown status code %d", resp.StatusCode())
	}
	if fmt.Sprint(order.ID) != ao.Order {
		log.Warn().Uint64("want", order.ID).Str("got", ao.Order).Msg("order ID not match")
		return nil
	}
	if order.Status == ao.Status && order.Status == "PROCESSING" {
		log.Debug().Uint64("order already in processing", order.ID)
		return nil
	}
	if !isValidStatus(ao.Status) {
		log.Warn().Str("unknown status detected", ao.Status)
		return nil
	}
	order.Status = ao.Status
	order.Accrual = uint64(ao.Accrual * 100)
	if err = bo.storage.UpdateOrder(order); err != nil {
		return fmt.Errorf("failed to update order ID %d - %s", order.ID, err.Error())
	}
	log.Debug().Uint64("successfully updated order", order.ID)
	return nil
}

func NewBlackbox(st entity.Storager, addr string) *Blackbox {
	return &Blackbox{
		limit:   1000,
		url:     addr + "/api/orders/",
		storage: st,
	}
}

func (b *Blackbox) Start() {
	rand.Seed(time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sig
		cancel()
	}()
	b.run(ctx)
}

func (b *Blackbox) run(ctx context.Context) {
	for {
		b.updatePool()

		g, _ := errgroup.WithContext(ctx)
		for _, order := range b.pool {
			w := &blackboxOrder{Blackbox: b, ctx: ctx, order: order}
			g.Go(w.Do)
		}
		err := g.Wait()
		if err != nil {
			atomic.StoreInt32(&b.needSleep, 1)
			if !errors.Is(err, ErrTooManyRequests) {
				atomic.StoreUint32(&b.limit, 1000)
			}
			log.Error().Err(err).Msg("blackbox service request failed")
		}
		sleep := 1 * time.Second
		if atomic.LoadInt32(&b.needSleep) == 0 {
			atomic.AddUint32(&b.limit, 1)
		} else {
			sleep = 60 * time.Second
			atomic.StoreInt32(&b.needSleep, 0)
		}
		log.Debug().Uint32("new limit", atomic.LoadUint32(&b.limit))
		log.Debug().Dur("pause", sleep)
		select {
		case <-ctx.Done():
			return
		case <-time.After(sleep):
		}
	}
}

func (b *Blackbox) updatePool() {
	limit := atomic.LoadUint32(&b.limit)
	ors, err := b.storage.GetPullOrders(limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to get orders for pool")
		return
	}
	count := uint32(0)
	pool := make(map[uint64]*entity.Order, limit)
	for k, order := range ors {
		count++
		if count > limit {
			break
		}
		pool[k] = order
	}
	b.pool = pool
	log.Debug().Int("length of pool", len(b.pool)).Msg("Orders pool updated, now in pool")
}

func isValidStatus(status string) bool {
	switch status {
	case "NEW":
		return true
	case "PROCESSING":
		return true
	case "PROCESSED":
		return true
	case "INVALID":
		return true
	default:
		return false
	}
}
