package adapters

import (
	"context"
	"sync"
)

type FilesAPIStub struct {
	lock        sync.Mutex
	PutCalls    []PutCall
	StoredFiles map[string]string
	// Track processed fileIDs for idempotency
	processedFiles map[string]struct{}
}

type PutCall struct {
	FileID      string
	FileContent string
}

func NewFilesAPIStub() *FilesAPIStub {
	return &FilesAPIStub{
		StoredFiles:    make(map[string]string),
		processedFiles: make(map[string]struct{}),
	}
}

func (s *FilesAPIStub) UploadFile(ctx context.Context, fileID, fileContent string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Check idempotency
	if _, exists := s.processedFiles[fileID]; exists {
		return nil
	}

	s.processedFiles[fileID] = struct{}{}
	s.PutCalls = append(s.PutCalls, PutCall{
		FileID:      fileID,
		FileContent: fileContent,
	})

	s.StoredFiles[fileID] = fileContent

	return nil
}

func (s *FilesAPIStub) FindPutCallByFileID(fileID string) (PutCall, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, call := range s.PutCalls {
		if call.FileID == fileID {
			return call, true
		}
	}
	return PutCall{}, false
}

func (s *FilesAPIStub) PutCallsCount() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.PutCalls)
}
