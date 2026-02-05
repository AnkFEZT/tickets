package adapters

import (
	"context"
	"sync"
)

type SpreadsheetsAPIStub struct {
	lock sync.Mutex
	Rows map[string][][]string
	// Track ticket IDs per sheet for idempotency (ticketID -> empty struct)
	processedTickets map[string]map[string]struct{}
}

func NewSpreadsheetsAPIStub() *SpreadsheetsAPIStub {
	return &SpreadsheetsAPIStub{
		processedTickets: make(map[string]map[string]struct{}),
	}
}

func (c *SpreadsheetsAPIStub) AppendRow(ctx context.Context, spreadsheetName string, row []string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.Rows == nil {
		c.Rows = make(map[string][][]string)
	}

	if c.processedTickets[spreadsheetName] == nil {
		c.processedTickets[spreadsheetName] = make(map[string]struct{})
	}

	// Use ticketID (first column) for idempotency check
	ticketID := row[0]
	if _, exists := c.processedTickets[spreadsheetName][ticketID]; exists {
		return nil // Already processed this ticket for this sheet
	}

	c.processedTickets[spreadsheetName][ticketID] = struct{}{}
	c.Rows[spreadsheetName] = append(c.Rows[spreadsheetName], row)

	return nil
}

func (c *SpreadsheetsAPIStub) FindRowByTicketID(sheetName, ticketID string) ([]string, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	rows, ok := c.Rows[sheetName]
	if !ok {
		return nil, false
	}

	for _, row := range rows {
		for _, col := range row {
			if col == ticketID {
				return row, true
			}
		}
	}
	return nil, false
}

func (c *SpreadsheetsAPIStub) HasSheet(sheetName string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.Rows[sheetName]
	return ok
}
