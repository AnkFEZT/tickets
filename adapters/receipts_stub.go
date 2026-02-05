package adapters

import (
	"context"
	"sync"
	"tickets/entities"
	"time"
)

type ReceiptsServiceStub struct {
	lock                sync.Mutex
	IssuedReceipts      []entities.IssueReceiptRequest
	VoidedReceipts      []VoidReceiptRequest
	usedIdempotencyKeys map[string]struct{}
	ReceiptResponses    map[string]entities.IssueReceiptResponse
}

type VoidReceiptRequest struct {
	TicketID       string
	IdempotencyKey string
}

func NewReceiptsServiceStub() *ReceiptsServiceStub {
	return &ReceiptsServiceStub{
		usedIdempotencyKeys: make(map[string]struct{}),
		ReceiptResponses:    make(map[string]entities.IssueReceiptResponse),
	}
}

func (s *ReceiptsServiceStub) IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) (entities.IssueReceiptResponse, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	idempotencyKey := request.IdempotencyKey + request.TicketID
	if _, exists := s.usedIdempotencyKeys[idempotencyKey]; exists {
		// Return the previously stored response for idempotency
		if resp, ok := s.ReceiptResponses[idempotencyKey]; ok {
			return resp, nil
		}
		return entities.IssueReceiptResponse{}, nil
	}

	s.usedIdempotencyKeys[idempotencyKey] = struct{}{}
	s.IssuedReceipts = append(s.IssuedReceipts, request)

	response := entities.IssueReceiptResponse{
		ReceiptNumber: "receipt-" + request.TicketID,
		IssuedAt:      time.Now(),
	}
	s.ReceiptResponses[idempotencyKey] = response

	return response, nil
}

func (s *ReceiptsServiceStub) VoidReceipt(ctx context.Context, ticketID, idempotencyKey string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, exists := s.usedIdempotencyKeys[idempotencyKey]; exists {
		return nil
	}

	s.usedIdempotencyKeys[idempotencyKey] = struct{}{}
	s.VoidedReceipts = append(s.VoidedReceipts, VoidReceiptRequest{
		TicketID:       ticketID,
		IdempotencyKey: idempotencyKey,
	})

	return nil
}

func (s *ReceiptsServiceStub) FindIssuedReceipt(ticketID string) (entities.IssueReceiptRequest, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, r := range s.IssuedReceipts {
		if r.TicketID == ticketID {
			return r, true
		}
	}
	return entities.IssueReceiptRequest{}, false
}

func (s *ReceiptsServiceStub) FindVoidedReceipt(ticketID string) (VoidReceiptRequest, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, r := range s.VoidedReceipts {
		if r.TicketID == ticketID {
			return r, true
		}
	}
	return VoidReceiptRequest{}, false
}

func (s *ReceiptsServiceStub) IssuedReceiptsCount() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.IssuedReceipts)
}

func (s *ReceiptsServiceStub) VoidedReceiptsCount() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.VoidedReceipts)
}
