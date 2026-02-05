package adapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

type FilesAPIClient struct {
	clients *clients.Clients
}

func NewFilesAPIClient(clients *clients.Clients) *FilesAPIClient {
	if clients == nil {
		panic("NewFilesAPIClient: clients is nil")
	}

	return &FilesAPIClient{clients: clients}
}

func (c FilesAPIClient) UploadFile(ctx context.Context, fileID, fileContent string) error {
	resp, err := c.clients.Files.PutFilesFileIdContentWithTextBodyWithResponse(ctx, fileID, fileContent)
	if err != nil {
		return fmt.Errorf("failed to put file content: %w", err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("failed to put file content: unexpected status code %d", resp.StatusCode())
	}

	if resp.StatusCode() == http.StatusConflict {
		log.FromContext(ctx).With("file", fileID).Info("file already exists")
		return nil
	}

	return nil
}
