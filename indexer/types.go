package indexer

import "sync"

type SyncStatus struct {
	latestBlockNum uint64
	lock           sync.RWMutex
}

func NewSyncStatus() *SyncStatus {
	return &SyncStatus{0, sync.RWMutex{}}
}

func (s *SyncStatus) Update(blockNum uint64) {
	s.lock.Lock()
	s.latestBlockNum = blockNum
	s.lock.Unlock()
}
