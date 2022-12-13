package service

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
)

type accrualOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type Accrl struct {
	url       string
	storage   Storager
	limit     uint32
	needSleep int32
	pool      map[uint64]*Order
}

type accOrder struct {
	*Accrl
	ctx   context.Context
	order *Order
}

func NewAccrl(s Storager, addr string) *Accrl {

	return &Accrl{
		limit:   1000,
		url:     addr + "/api/orders/",
		storage: s,
	}
}

func (a *Accrl) Run() {
	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// штатное завершение по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sig
		cancel()
	}()

	a.start(ctx)
}

func (a *Accrl) start(ctx context.Context) {
	for {
		a.refresh()

		g, _ := errgroup.WithContext(ctx) // используем errgroup
		for _, order := range a.pool {
			w := &accOrder{Accrl: a, ctx: ctx, order: order}
			g.Go(w.Do)
		}
		err := g.Wait()
		if err != nil {
			atomic.StoreInt32(&a.needSleep, 1) // случилась ошибка, выставим флаг сделать паузу
			if !errors.Is(err, ErrTooManyRequests) {
				// если это не ошибка с превышением кол-ва запросов, выставим лимит по умолчанию
				// иначе, лимит уже был выставлен после парсинга ответа
				atomic.StoreUint32(&a.limit, 1000)
			}
			msg := fmt.Sprintf("accrual service request failed -%s", err.Error())
			log.Info().Msg(msg)
		}

		sleep := 1 * time.Second // дадим секундную передышку сервису `accrual`
		if atomic.LoadInt32(&a.needSleep) == 0 {
			// так как делать большую паузу не нужно, увеличим лимит возможных запросов
			atomic.AddUint32(&a.limit, 1)
		} else {
			// воркер столкнулся с ошибкой или был превышен лимит, сделаем паузу на минуту
			sleep = 60 * time.Second
			// новый лимит уже был выставлен воркером, первым столкнувшимся с ошибкой
			// поэтому просто обнулим флаг `needSleep`
			atomic.StoreInt32(&a.needSleep, 0)
		}
		msg := fmt.Sprint("got new limit:", atomic.LoadUint32(&a.limit))
		log.Info().Msg(msg)
		msg = fmt.Sprintf("sleeping for %s seconds", sleep)
		log.Info().Msg(msg)
		select {
		case <-ctx.Done():
			return
		case <-time.After(sleep):
		}
	}
}

func (a *Accrl) refresh() {
	limit := atomic.LoadUint32(&a.limit)

	ors, err := a.storage.PullOrders(limit) // получаем заказы со статусом NEW и PROCESSING, отсортированные по дате поступления
	if err != nil {
		msg := fmt.Sprintf("failed to get orders for pool -", err)
		log.Info().Msg(msg)
		return
	}

	count := uint32(0)
	pool := make(map[uint64]*Order, limit)

	for k, order := range ors {
		count++
		if count > limit {
			break
		}
		pool[k] = order
	}

	a.pool = pool
	msg := fmt.Sprintf("orders pool updated, now in pool: %d", len(a.pool))
	log.Info().Msg(msg)
}

func (ao *accOrder) Do() error {
	ctx, cancel := context.WithTimeout(ao.ctx, 60*time.Second)
	defer cancel()
	order := ao.order
	url := fmt.Sprintf("%s%d", ao.url, order.ID)
	msg := fmt.Sprintf("make request:%s", url)
	log.Info().Msg(msg)

	a := &accrualOrder{}
	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", "*/*").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Length", "0").
		SetContext(ctx).
		SetResult(&a).
		Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusInternalServerError {
		return fmt.Errorf("internal server error, status code %d", resp.StatusCode())
	}

	if resp.StatusCode() == http.StatusTooManyRequests {
		n := uint32(rand.Intn(10)) + 2 // эмулятор переменного кол-во запросов
		atomic.StoreUint32(&ao.limit, n)
		msg = fmt.Sprintf("too many requests detected", string(resp.Body()))
		log.Info().Msg(msg)
		return ErrTooManyRequests
	}

	if resp.StatusCode() == http.StatusNoContent {
		// некритичная ошибка, отменять выполнение других воркеров не надо: выведем варнинг и выйдем из рутины
		msg = fmt.Sprintf("no content for order %d\n", order.ID)
		log.Info().Msg(msg)
		return nil
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unknown status code %d", resp.StatusCode())
	}

	if fmt.Sprint(order.ID) != a.Order {
		// некритичная ошибка
		msg = fmt.Sprintf("order ID not match, want %d, got %s\n", order.ID, a.Order)
		log.Info().Msg(msg)
		return nil
	}

	if order.Status == a.Status && order.Status == "PROCESSING" {
		// заказ уже находится в обработке, выходим
		msg = fmt.Sprintf("order %d already in processing\n", order.ID)
		log.Info().Msg(msg)
		return nil
	}

	if !isValidStatus(a.Status) {
		//некритичная ошибка
		msg = fmt.Sprintf("unknown status detected: %s\n", a.Status)
		log.Info().Msg(msg)
		return nil
	}

	order.Status = a.Status
	order.Accrual = uint64(a.Accrual * 100)

	// запрос успешно выполнен, обновим заказ
	if err = ao.storage.UpdateOrder(order); err != nil {
		return fmt.Errorf("failed to update order ID %d - %w", order.ID, err)
	}
	msg = fmt.Sprintf("order successfully updated: order %v\n", order)
	log.Info().Msg(msg)
	return nil
}

func isValidStatus(status string) bool {
	switch status {
	case "NEW":
	case "PROCESSING":
	case "PROCESSED":
	case "INVALID":
	default:
		return false
	}

	return true
}
