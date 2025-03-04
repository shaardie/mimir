// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/thanos-io/thanos/blob/main/pkg/store/bucket.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Thanos Authors.

package storegateway

import (
	"sync"
	"time"
)

// queryStats holds query statistics. This data structure is NOT concurrency safe.
type queryStats struct {
	blocksQueried int

	postingsTouched          int
	postingsTouchedSizeSum   int
	postingsToFetch          int
	postingsFetched          int
	postingsFetchedSizeSum   int
	postingsFetchCount       int
	postingsFetchDurationSum time.Duration

	cachedPostingsCompressions         int
	cachedPostingsCompressionErrors    int
	cachedPostingsOriginalSizeSum      int
	cachedPostingsCompressedSizeSum    int
	cachedPostingsCompressionTimeSum   time.Duration
	cachedPostingsDecompressions       int
	cachedPostingsDecompressionErrors  int
	cachedPostingsDecompressionTimeSum time.Duration

	seriesProcessed        int
	seriesProcessedSizeSum int
	seriesOmitted          int
	seriesFetched          int
	seriesFetchedSizeSum   int
	seriesFetchDurationSum time.Duration
	seriesRefetches        int

	seriesHashCacheRequests int
	seriesHashCacheHits     int

	chunksTouched          int
	chunksTouchedSizeSum   int
	chunksFetched          int
	chunksFetchedSizeSum   int
	chunksRefetched        int
	chunksRefetchedSizeSum int
	chunksProcessed        int
	chunksProcessedSizeSum int
	chunksReturned         int
	chunksReturnedSizeSum  int

	mergedSeriesCount int
	mergedChunksCount int

	// The total time spent fetching series and chunk refs.
	streamingSeriesFetchRefsDuration time.Duration

	// The number of batches the Series() request has been split into.
	streamingSeriesBatchCount int

	// The total time spent loading batches.
	streamingSeriesBatchLoadDuration time.Duration

	// The total time spent waiting until the next batch is loaded, once the store-gateway was
	// ready to send it to the client.
	streamingSeriesWaitBatchLoadedDuration time.Duration

	// The Series() request timing breakdown.
	streamingSeriesExpandPostingsDuration       time.Duration
	streamingSeriesFetchSeriesAndChunksDuration time.Duration
	streamingSeriesEncodeResponseDuration       time.Duration
	streamingSeriesSendResponseDuration         time.Duration
	streamingSeriesOtherDuration                time.Duration
	streamingSeriesIndexHeaderLoadDuration      time.Duration
}

func (s queryStats) merge(o *queryStats) *queryStats {
	s.blocksQueried += o.blocksQueried

	s.postingsTouched += o.postingsTouched
	s.postingsTouchedSizeSum += o.postingsTouchedSizeSum
	s.postingsToFetch += o.postingsToFetch
	s.postingsFetched += o.postingsFetched
	s.postingsFetchedSizeSum += o.postingsFetchedSizeSum
	s.postingsFetchCount += o.postingsFetchCount
	s.postingsFetchDurationSum += o.postingsFetchDurationSum

	s.cachedPostingsCompressions += o.cachedPostingsCompressions
	s.cachedPostingsCompressionErrors += o.cachedPostingsCompressionErrors
	s.cachedPostingsOriginalSizeSum += o.cachedPostingsOriginalSizeSum
	s.cachedPostingsCompressedSizeSum += o.cachedPostingsCompressedSizeSum
	s.cachedPostingsCompressionTimeSum += o.cachedPostingsCompressionTimeSum
	s.cachedPostingsDecompressions += o.cachedPostingsDecompressions
	s.cachedPostingsDecompressionErrors += o.cachedPostingsDecompressionErrors
	s.cachedPostingsDecompressionTimeSum += o.cachedPostingsDecompressionTimeSum

	s.seriesProcessed += o.seriesProcessed
	s.seriesProcessedSizeSum += o.seriesProcessedSizeSum
	s.seriesOmitted += o.seriesOmitted
	s.seriesFetched += o.seriesFetched
	s.seriesFetchedSizeSum += o.seriesFetchedSizeSum
	s.seriesFetchDurationSum += o.seriesFetchDurationSum
	s.seriesRefetches += o.seriesRefetches

	s.seriesHashCacheRequests += o.seriesHashCacheRequests
	s.seriesHashCacheHits += o.seriesHashCacheHits

	s.chunksTouched += o.chunksTouched
	s.chunksTouchedSizeSum += o.chunksTouchedSizeSum
	s.chunksFetched += o.chunksFetched
	s.chunksFetchedSizeSum += o.chunksFetchedSizeSum
	s.chunksRefetched += o.chunksRefetched
	s.chunksRefetchedSizeSum += o.chunksRefetchedSizeSum
	s.chunksProcessed += o.chunksProcessed
	s.chunksProcessedSizeSum += o.chunksProcessedSizeSum
	s.chunksReturned += o.chunksReturned
	s.chunksReturnedSizeSum += o.chunksReturnedSizeSum

	s.mergedSeriesCount += o.mergedSeriesCount
	s.mergedChunksCount += o.mergedChunksCount

	s.streamingSeriesFetchRefsDuration += o.streamingSeriesFetchRefsDuration
	s.streamingSeriesBatchCount += o.streamingSeriesBatchCount
	s.streamingSeriesBatchLoadDuration += o.streamingSeriesBatchLoadDuration
	s.streamingSeriesWaitBatchLoadedDuration += o.streamingSeriesWaitBatchLoadedDuration
	s.streamingSeriesExpandPostingsDuration += o.streamingSeriesExpandPostingsDuration
	s.streamingSeriesFetchSeriesAndChunksDuration += o.streamingSeriesFetchSeriesAndChunksDuration
	s.streamingSeriesEncodeResponseDuration += o.streamingSeriesEncodeResponseDuration
	s.streamingSeriesSendResponseDuration += o.streamingSeriesSendResponseDuration
	s.streamingSeriesOtherDuration += o.streamingSeriesOtherDuration

	s.streamingSeriesIndexHeaderLoadDuration += o.streamingSeriesIndexHeaderLoadDuration

	return &s
}

// safeQueryStats wraps queryStats adding functions manipulate the statistics while holding a lock.
type safeQueryStats struct {
	unsafeStatsMx sync.Mutex
	unsafeStats   *queryStats
}

func newSafeQueryStats() *safeQueryStats {
	return &safeQueryStats{
		unsafeStats: &queryStats{},
	}
}

// update the statistics while holding the lock.
func (s *safeQueryStats) update(fn func(stats *queryStats)) {
	s.unsafeStatsMx.Lock()
	defer s.unsafeStatsMx.Unlock()

	fn(s.unsafeStats)
}

// merge the statistics while holding the lock. Statistics are merged in the receiver.
func (s *safeQueryStats) merge(o *queryStats) {
	s.unsafeStatsMx.Lock()
	defer s.unsafeStatsMx.Unlock()

	s.unsafeStats = s.unsafeStats.merge(o)
}

// export returns a copy of the internal statistics.
func (s *safeQueryStats) export() *queryStats {
	s.unsafeStatsMx.Lock()
	defer s.unsafeStatsMx.Unlock()

	copied := *s.unsafeStats
	return &copied
}
