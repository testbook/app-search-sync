package main

import "time"

type bulkProcessorStats struct {
	Enabled      bool
	Flushed      int64 // number of times the flush interval has been invoked
	Committed    int64 // # of times workers committed bulk requests
	Indexed      int64 // # of requests indexed
	Created      int64 // # of requests that ES reported as creates (201)
	Updated      int64 // # of requests that ES reported as updates
	Deleted      int64 // # of requests that ES reported as deletes
	Succeeded    int64 // # of requests that ES reported as successful
	Failed       int64 // # of requests that ES reported as failed
	LastUpdateTs time.Time
}

func (s *bulkProcessorStats) AddFlushed(c int) {
	if !s.Enabled {
		return
	}
	s.Flushed += int64(c)
	s.LastUpdateTs = time.Now()
}

func (s *bulkProcessorStats) AddCommitted(c int) {
	if !s.Enabled {
		return
	}
	s.Committed += int64(c)
	s.LastUpdateTs = time.Now()
}

func (s *bulkProcessorStats) AddIndexed(c int) {
	if !s.Enabled {
		return
	}
	s.Indexed += int64(c)
	s.LastUpdateTs = time.Now()
}

func (s *bulkProcessorStats) AddCreated(c int) {
	if !s.Enabled {
		return
	}
	s.Created += int64(c)
	s.LastUpdateTs = time.Now()
}

func (s *bulkProcessorStats) AddUpdated(c int) {
	if !s.Enabled {
		return
	}
	s.Updated += int64(c)
	s.LastUpdateTs = time.Now()
}

func (s *bulkProcessorStats) AddDeleted(c int) {
	if !s.Enabled {
		return
	}
	s.Deleted += int64(c)
	s.LastUpdateTs = time.Now()
}

func (s *bulkProcessorStats) AddSucceeded(c int) {
	if !s.Enabled {
		return
	}
	s.Succeeded += int64(c)
	s.LastUpdateTs = time.Now()
}

func (s *bulkProcessorStats) AddFailed(c int) {
	if !s.Enabled {
		return
	}
	s.Failed += int64(c)
	s.LastUpdateTs = time.Now()
}
