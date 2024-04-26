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
	"github.com/KretovDmitry/gophermart/pkg/limiter"
	"github.com/KretovDmitry/gophermart/pkg/logger"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/shopspring/decimal"
)

type AccrualService struct {
	orderRepo   repositories.OrderRepository
	accountRepo repositories.AccountRepository
	trm         *manager.Manager
	logger      logger.Logger
	config      *config.Config
	client      *http.Client
	limiter     *limiter.DynamicRateLimiter
	wg          *sync.WaitGroup
	done        chan struct{}
}

func NewAccrualService(
	orderRepo repositories.OrderRepository,
	accountRepo repositories.AccountRepository,
	trm *manager.Manager,
	config *config.Config,
	logger logger.Logger,
) (*AccrualService, error) {
	if config == nil {
		return nil, errors.New("nil dependency: config")
	}
	if trm == nil {
		return nil, errors.New("nil dependency: transaction manager")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create new cookie jar: %w", err)
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: config.Accrual.Timeout,
	}

	limiter := limiter.NewDynamicRateLimiter(config.Accrual.Every, config.Accrual.Burst)

	return &AccrualService{
		orderRepo:   orderRepo,
		accountRepo: accountRepo,
		trm:         trm,
		logger:      logger,
		config:      config,
		client:      client,
		limiter:     limiter,
		wg:          &sync.WaitGroup{},
		done:        make(chan struct{}),
	}, nil
}

func (s *AccrualService) Run(ctx context.Context) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run(ctx)
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

func (s *AccrualService) run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	offset := 0

	for {
		select {
		case <-s.done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.logger.Debugf("offset: %d", offset)

			orders, err := s.orderRepo.GetUnprocessedOrders(ctx, s.config.Accrual.Limit, offset)
			if err != nil {
				if errors.Is(err, errs.ErrNotFound) {
					offset = 0
					continue
				}
				s.logger.Errorf("get unprocessed orders: %v", err)
				continue
			}

			offset += len(orders)

			for _, ord := range orders {
				ord := ord

				if err = s.limiter.Wait(ctx); err != nil {
					if !errors.Is(err, context.Canceled) {
						s.logger.Errorf("wait limiter: %v", err)
					}
				}

				if err = s.update(ctx, ord.Number); err != nil {
					if errors.Is(err, errs.ErrRateLimit) {
						time.Sleep(time.Minute)
						s.limiter.Update(s.config.Accrual.Every+time.Second, s.config.Accrual.Burst)
						continue
					}
				}
			}
		}
	}
}

func (s *AccrualService) update(ctx context.Context, num entities.OrderNumber) error {
	info, err := s.get(ctx, num)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("get order info: %w", err)
	}

	s.logger.Debugf("got order info: %v", info)

	return s.trm.Do(ctx, func(ctx context.Context) error {
		userID, err := s.orderRepo.UpdateOrder(ctx, info)
		if err != nil {
			return fmt.Errorf("update order: %w", err)
		}

		if info.Accrual.GreaterThan(decimal.NewFromInt(0)) {
			if err = s.accountRepo.AddToAccount(ctx, userID, info.Accrual); err != nil {
				return fmt.Errorf("add to account: %w", err)
			}
		}

		return nil
	})
}

func (s *AccrualService) get(ctx context.Context, num entities.OrderNumber) (*entities.UpdateOrderInfo, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.config.Accrual.Address, num)

	s.logger.Debugf("doing request for: %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	switch res.StatusCode {
	case http.StatusTooManyRequests:
		return nil, errs.ErrRateLimit
	case http.StatusNoContent:
		s.logger.Debugf("no data for order: %q", num)
		return nil, errs.ErrNotFound
	}

	payload := new(accrual.UpdateOrderInfo)

	defer res.Body.Close()

	if err = json.NewDecoder(res.Body).Decode(payload); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return entities.NewUpdateInfoFromResponse(payload), nil
}
