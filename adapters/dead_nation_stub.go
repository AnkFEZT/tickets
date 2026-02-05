package adapters

import (
	"context"
	"sync"

	"tickets/entities"
)

type DeadNationStub struct {
	lock        sync.Mutex
	Bookings    []entities.DeadNationBooking
	ShouldError error
}

func NewDeadNationStub() *DeadNationStub {
	return &DeadNationStub{}
}

func (s *DeadNationStub) BookTicket(ctx context.Context, request entities.DeadNationBooking) error {
	if s.ShouldError != nil {
		return s.ShouldError
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.Bookings = append(s.Bookings, request)

	return nil
}

func (s *DeadNationStub) FindBookingByBookingID(bookingID string) (entities.DeadNationBooking, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, booking := range s.Bookings {
		if booking.BookingID.String() == bookingID {
			return booking, true
		}
	}
	return entities.DeadNationBooking{}, false
}

func (s *DeadNationStub) BookingsCount() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.Bookings)
}
