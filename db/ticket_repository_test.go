package db_test

import (
	"context"
	"os"
	"testing"
	"tickets/db"
	"tickets/entities"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testDB := getDBTest()

	if err := db.InitializeDatabaseSchema(testDB); err != nil {
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}

func TestTicketsRepository_Add_idempotency(t *testing.T) {
	ctx := context.Background()
	repo := db.NewTicketRepository(getDBTest())

	ticketToAdd := entities.Ticket{
		TicketID: uuid.NewString(),
		Price: entities.Money{
			Amount:   "30.00",
			Currency: "EUR",
		},
		CustomerEmail: "foo@bar.com",
	}

	for i := 0; i < 2; i++ {
		err := repo.Add(ctx, ticketToAdd)
		require.NoError(t, err)

		tickets, err := repo.FindAll(ctx)
		require.NoError(t, err)

		foundTickets := lo.Filter(tickets, func(t entities.Ticket, _ int) bool {
			return t.TicketID == ticketToAdd.TicketID
		})

		// add should be idempotent, so the method should always return 1
		require.Len(t, foundTickets, 1)
	}
}
