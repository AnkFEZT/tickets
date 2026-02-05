package http

import (
	"tickets/db"

	libHttp "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func NewHttpRouter(eventBus *cqrs.EventBus, commandBus *cqrs.CommandBus, tickets db.TicketRepository, shows db.ShowRepository, bookings db.BookingRepository, opsBookings db.OpsBookingReadModel) *echo.Echo {
	e := libHttp.NewEcho()

	e.Use(otelecho.Middleware("tickets"))
	e.Use(TraceIDMiddleware())

	handler := Handler{
		eventBus:    eventBus,
		commandBus:  commandBus,
		tickets:     tickets,
		shows:       shows,
		bookings:    bookings,
		opsBookings: opsBookings,
	}

	e.POST("/tickets-status", handler.PostTicketsStatus)
	e.GET("/tickets", handler.GetTickets)
	e.POST("/shows", handler.PostShows)
	e.POST("/book-tickets", handler.PostBookTickets)
	e.PUT("/ticket-refund/:ticket_id", handler.PutTicketRefund)
	e.GET("/health", handler.GetHealthCheck)

	e.GET("/ops/bookings", handler.GetOpsBookings)
	e.GET("/ops/bookings/:id", handler.GetOpsBookingByID)

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	return e
}
