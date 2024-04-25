package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"github.com/KretovDmitry/gophermart/internal/application/errs"
	"github.com/KretovDmitry/gophermart/internal/config"
	"github.com/KretovDmitry/gophermart/internal/domain/entities"
	"github.com/KretovDmitry/gophermart/internal/domain/repositories"
	"github.com/KretovDmitry/gophermart/internal/interface/api/rest/response/accrual"
	"github.com/KretovDmitry/gophermart/pkg/logger"
)

type AccrualService struct {
	orderRepo   repositories.OrderRepository
	accountRepo repositories.AccountRepository
	logger      logger.Logger
	config      *config.Config
	client      *http.Client
	wg          *sync.WaitGroup
	done        chan struct{}
}

func NewAccrualService(
	orderRepo repositories.OrderRepository,
	accountRepo repositories.AccountRepository,
	config *config.Config,
	logger logger.Logger,
) (*AccrualService, error) {
	if config == nil {
		return nil, errors.New("nil dependency: config")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create new cookie jar: %w", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: config.Accrual.Timeout,
	}

	return &AccrualService{
		orderRepo:   orderRepo,
		accountRepo: accountRepo,
		logger:      logger,
		config:      config,
		client:      client,
		wg:          &sync.WaitGroup{},
		done:        make(chan struct{}),
	}, nil
}

func (s *AccrualService) Run() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run()
	}()
}

func (s *AccrualService) Stop() {
	sync.OnceFunc(func() {
		close(s.done)
	})()

	ready := make(chan struct{})
	go func() {
		defer close(ready)
		s.wg.Wait()
	}()

	select {
	case <-time.After(s.config.HTTPServer.ShutdownTimeout):
		s.logger.Error("accrual service stop: shutdown timeout exceeded")
	case <-ready:
		return
	}
}

func (s *AccrualService) run() {
	offset := 0

	for {
		select {
		case <-s.done:
			return
		default:
			nums, err := s.orderRepo.GetUnprocessedOrderNumbers(
				context.TODO(),
				s.config.Accrual.Limit,
				offset)
			if err != nil {
				s.logger.Error(err)
				time.Sleep(1 * time.Minute)
				continue
			}
			if err := s.fun(nums...); err != nil {
				if errors.Is(err, errs.ErrRateLimit) {
					time.Sleep(1 * time.Minute)
					continue
				}
				// TODO: ErrNotFound reset offset
			}
			offset += s.config.Accrual.Limit
		}
	}
}

func (s *AccrualService) fun(nums ...entities.OrderNumber) error {
	var wg sync.WaitGroup

	res := make(chan entities.UpdateOrderInfo, s.config.Accrual.Limit)

	wg.Add(1)
	go func() {
		semaphore := make(chan struct{}, 5)

		// have a max rate of 10/sec
		rate := make(chan struct{}, 10)
		for i := 0; i < cap(rate); i++ {
			rate <- struct{}{}
		}

		// leaky bucket
		go func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for range ticker.C {
				_, ok := <-rate
				if !ok {
					return
				}
			}
		}()

		var wg sync.WaitGroup
		for _, num := range nums {
			num := num
			wg.Add(1)
			go func() {
				defer wg.Done()

				// wait for the rate limiter
				rate <- struct{}{}

				// check the concurrency semaphore
				semaphore <- struct{}{}
				defer func() {
					<-semaphore
				}()

				go s.get(num, res)
			}()
		}
		wg.Wait()
		close(rate)
	}()
	wg.Wait()

	return nil
}

func (s *AccrualService) get(num entities.OrderNumber, out chan entities.UpdateOrderInfo) error {
	url := fmt.Sprintf("%s/api/orders/%s", s.config.Accrual.Address, num)

	res, err := s.client.Get(url)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusTooManyRequests {
		return errs.ErrRateLimit
	}

	payload := new(accrual.UpdateOrderInfo)

	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(payload); err != nil {
		return err
	}

	out <- *entities.NewUpdateInfoFromResponse(payload)

	return nil
}
