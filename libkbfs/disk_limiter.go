package libkbfs

import (
	"github.com/keybase/client/go/protocol/keybase1"
	"golang.org/x/net/context"
)

type diskLimitTrackerType int

const (
	unknownLimitTracker diskLimitTrackerType = iota
	diskCacheLimitTracker
	syncCacheLimitTracker
)

type diskLimitByteTracker interface {
	onEnable(usedResources int64)
	onDisable(usedResources int64)
	updateFree(freeResources int64)
	reserve(ctx context.Context, resources int64) (available int64, err error)
	tryReserve(resources int64) (available int64)
	commit(resources int64)
	rollback(resources int64)
	commitOrRollback(resources int64, shouldCommit bool)
	releaseAndCommit(resources int64)
}

type diskBlockCacheLimiter interface {
	// beforeDiskBlockCachePut is called by the disk block cache before putting
	// a block into the cache. It returns the total number of available bytes.
	beforeDiskBlockCachePut(ctx context.Context, blockBytes int64) (
		availableBytes int64, err error)

	// afterDiskBlockCachePut is called by the disk block cache after putting
	// a block into the cache. It returns how many bytes it acquired.
	afterDiskBlockCachePut(ctx context.Context, blockBytes int64,
		putData bool)
}

// DiskLimiter is an interface for limiting disk usage.
type DiskLimiter interface {
	diskBlockCacheLimiter

	// onByteTrackerEnable is called when a byte tracker is enabled to begin
	// accounting.
	onByteTrackerEnable(ctx context.Context, cacheBytes int64)

	// onByteTrackerDisable is called when a byte tracker is disable to stop
	// accounting.
	onByteTrackerDisable(ctx context.Context, cacheBytes int64)

	// onJournalEnable is called when initializing a TLF journal
	// with that journal's current disk usage. Both journalBytes
	// and journalFiles must be >= 0. The updated available byte
	// and file count must be returned.
	onJournalEnable(ctx context.Context,
		journalStoredBytes, journalUnflushedBytes, journalFiles int64,
		chargedTo keybase1.UserOrTeamID) (
		availableBytes, availableFiles int64)

	// onJournalDisable is called when shutting down a TLF journal
	// with that journal's current disk usage. Both journalBytes
	// and journalFiles must be >= 0.
	onJournalDisable(ctx context.Context,
		journalStoredBytes, journalUnflushedBytes, journalFiles int64,
		chargedTo keybase1.UserOrTeamID)

	// beforeBlockPut is called before putting a block of the
	// given byte and file count, both of which must be > 0. It
	// may block, but must return immediately with a
	// (possibly-wrapped) ctx.Err() if ctx is cancelled. The
	// updated available byte and file count must be returned,
	// even if err is non-nil.
	beforeBlockPut(ctx context.Context,
		blockBytes, blockFiles int64, chargedTo keybase1.UserOrTeamID) (
		availableBytes, availableFiles int64, err error)

	// afterBlockPut is called after putting a block of the given
	// byte and file count, which must match the corresponding call to
	// beforeBlockPut. putData reflects whether or not the data
	// was actually put; if it's false, it's either because of an
	// error or because the block already existed.
	afterBlockPut(ctx context.Context,
		blockBytes, blockFiles int64, putData bool,
		chargedTo keybase1.UserOrTeamID)

	// onBlocksFlush is called after flushing blocks of the given
	// byte count, which must be >= 0. (Flushing a block with a
	// zero byte count shouldn't happen, but may as well let it go
	// through.)
	onBlocksFlush(ctx context.Context, blockBytes int64,
		chargedTo keybase1.UserOrTeamID)

	// onBlocksDelete is called after deleting blocks for a given tracker of
	// the given byte and file count, both of which must be >= 0.
	onBlocksDelete(ctx context.Context, typ diskLimitTrackerType, blockBytes,
		blockFiles int64)

	// getQuotaInfo returns the quota info as known by the disk
	// limiter.
	getQuotaInfo(chargedTo keybase1.UserOrTeamID) (
		usedQuotaBytes, quotaBytes int64)

	// getDiskLimitInfo returns the usage and limit info for the disk,
	// as known by the disk limiter.
	getDiskLimitInfo() (usedBytes int64, limitBytes float64,
		usedFiles int64, limitFiles float64)

	// getStatus returns an object that's marshallable into JSON
	// for use in displaying status.
	getStatus(chargedTo keybase1.UserOrTeamID) interface{}
}
