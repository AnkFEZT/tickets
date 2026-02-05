package adapters

import (
	"context"
	"sync"
)

type PaymentsServiceStub struct {
	lock                sync.Mutex
	RefundedPayments    []RefundPaymentRequest
	usedIdempotencyKeys map[string]struct{}
}

type RefundPaymentRequest struct {
	TicketID       string
	IdempotencyKey string
}

func NewPaymentsServiceStub() *PaymentsServiceStub {
	return &PaymentsServiceStub{
		usedIdempotencyKeys: make(map[string]struct{}),
	}
}

func (s *PaymentsServiceStub) RefundPayment(ctx context.Context, ticketID, idempotencyKey string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, exists := s.usedIdempotencyKeys[idempotencyKey]; exists {
		return nil
	}

	s.usedIdempotencyKeys[idempotencyKey] = struct{}{}
	s.RefundedPayments = append(s.RefundedPayments, RefundPaymentRequest{
		TicketID:       ticketID,
		IdempotencyKey: idempotencyKey,
	})

	return nil
}

func (s *PaymentsServiceStub) FindRefundedPayment(ticketID string) (RefundPaymentRequest, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, p := range s.RefundedPayments {
		if p.TicketID == ticketID {
			return p, true
		}
	}
	return RefundPaymentRequest{}, false
}

func (s *PaymentsServiceStub) RefundedPaymentsCount() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.RefundedPayments)
}
