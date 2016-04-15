package controller

import (
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"

	"github.com/giantswarm/inago/task"
)

// executeTaskAction executes the given function f to create a new task and block until
// it finishes. If the task fails, an error is raised.
func (c controller) executeTaskAction(f func(ctx context.Context, req Request) (*task.Task, error), ctx context.Context, req Request) error {
	taskObject, err := f(ctx, req)
	if err != nil {
		return maskAny(err)
	}
	closer := make(<-chan struct{})
	taskObject, err = c.WaitForTask(ctx, taskObject.ID, closer)
	if err != nil {
		return maskAny(err)
	}
	if task.HasFailedStatus(taskObject) {
		return maskAny(taskObject.Error)
	}
	return nil
}

func (c controller) getNumRunningSlices(ctx context.Context, req Request) (int, error) {
	c.Config.Logger.Debug(ctx, "controller: getting number of running slices")

	groupStatus, err := c.groupStatus(ctx, req)
	if IsUnitNotFound(err) {
		return 0, nil
	} else if err != nil {
		return 0, maskAny(err)
	}
	grouped, err := UnitStatusList(groupStatus).Group()
	if err != nil {
		return 0, maskAny(err)
	}

	c.Config.Logger.Debug(ctx, "controller: grouped list: %#v", grouped)

	var sliceIDs []string
	for _, us := range grouped {
		groupedStatuses := grouped.unitStatusesBySliceID(us.SliceID)
		if len(groupedStatuses) > 1 {
			// This group has an inconsistent state. Thus we do not consider it
			// running.
			continue
		}

		aggregator := Aggregator{
			Logger: c.Config.Logger,
		}
		ok, err := aggregator.UnitHasStatus(groupedStatuses[0], StatusRunning)
		if err != nil {
			return 0, maskAny(err)
		}
		if !ok {
			continue
		}

		sliceIDs = append(sliceIDs, us.SliceID)
	}

	c.Config.Logger.Debug(ctx, "controller: found %v running slices", len(sliceIDs))

	return len(sliceIDs), nil
}

func (c controller) isGroupRemovalAllowed(ctx context.Context, req Request, minAlive int, removeInProgress *int64) (bool, error) {
	c.Config.Logger.Debug(ctx, "controller: checking group removal allowed, req: %v", req)

	c.Config.Logger.Debug(
		ctx, "controller: removeInProgress: %v, minAlive: %v",
		*removeInProgress, minAlive,
	)

	if (minAlive - int(*removeInProgress)) > 0 {
		c.Config.Logger.Debug(ctx, "controller: group removal allowed ((minAlive - int(removeInProgress)) > 0)")
		return true, nil
	}

	c.Config.Logger.Debug(ctx, "controller: group removal not allowed ((minAlive - int(removeInProgress)) <= 0)")
	return false, nil
}

func (c controller) isGroupAdditionAllowed(ctx context.Context, req Request, maxGrowth int, additionInProgress *int64) (bool, error) {
	c.Config.Logger.Debug(ctx, "controller: checking group addition allowed, req: %v", req)

	c.Config.Logger.Debug(
		ctx, "controller: additionInProgress: %v, maxGrowth: %v",
		*additionInProgress, maxGrowth,
	)

	if (maxGrowth - int(*additionInProgress)) > 0 {
		c.Config.Logger.Debug(ctx, "controller: group addition allowed ((maxGrowth  - additionInProgress) > 0)")
		return true, nil
	}

	c.Config.Logger.Debug(ctx, "controller: group addition not allowed ((maxGrowth  - additionInProgress) <= 0)")
	return false, nil
}

func (c controller) addFirst(ctx context.Context, req Request, opts UpdateOptions) ([]string, error) {
	c.Config.Logger.Debug(ctx, "controller: running addFirst")

	c.Config.Logger.Debug(ctx, "controller: running add worker")
	newReq, err := c.runAddWorker(ctx, req, opts)
	if err != nil {
		return nil, maskAny(err)
	}

	c.Config.Logger.Debug(ctx, "controller: checking number of running slices")
	n, err := c.getNumRunningSlices(ctx, newReq)
	if err != nil {
		return nil, maskAny(err)
	}
	if n != len(newReq.SliceIDs) {
		return nil, maskAnyf(updateFailedError, "addFirst: slice not running: %d != %v", n, newReq.SliceIDs)
	}

	c.Config.Logger.Debug(ctx, "controller: running remove worker")
	err = c.runRemoveWorker(ctx, req)
	if err != nil {
		return nil, maskAny(err)
	}

	return newReq.SliceIDs, nil
}

func (c controller) runAddWorker(ctx context.Context, req Request, opts UpdateOptions) (Request, error) {
	c.Config.Logger.Info(ctx, "controller: adding units")

	// Create new random IDs.
	req.DesiredSlices = 1
	req.SliceIDs = nil
	newReq, err := c.ExtendWithRandomSliceIDs(ctx, req)
	if err != nil {
		return Request{}, maskAny(err)
	}

	// Submit.
	if err := c.executeTaskAction(c.Submit, ctx, newReq); err != nil {
		return Request{}, maskAny(err)
	}

	// Start.
	if err := c.executeTaskAction(c.Start, ctx, newReq); err != nil {
		return Request{}, maskAny(err)
	}

	time.Sleep(time.Duration(opts.ReadySecs) * time.Second)

	return newReq, nil
}

func (c controller) removeFirst(ctx context.Context, req Request, opts UpdateOptions) ([]string, error) {
	c.Config.Logger.Debug(ctx, "controller: running removeFirst")

	c.Config.Logger.Debug(ctx, "controller: running remove worker")
	err := c.runRemoveWorker(ctx, req)
	if err != nil {
		return nil, maskAny(err)
	}

	c.Config.Logger.Debug(ctx, "controller: running add worker")
	newReq, err := c.runAddWorker(ctx, req, opts)
	if err != nil {
		return nil, maskAny(err)
	}

	c.Config.Logger.Debug(ctx, "controller: checking number of running slices")
	n, err := c.getNumRunningSlices(ctx, newReq)
	if err != nil {
		return nil, maskAny(err)
	}
	if n != len(newReq.SliceIDs) {
		return nil, maskAnyf(updateFailedError, "removeFirst: slice not running: %d != %v", n, newReq.SliceIDs)
	}

	return newReq.SliceIDs, nil
}

func (c controller) runRemoveWorker(ctx context.Context, req Request) error {
	c.Config.Logger.Info(ctx, "controller: removing units")
	c.Config.Logger.Debug(ctx, "controller: executing stop action, req: %v", req)
	// Stop.
	if err := c.executeTaskAction(c.Stop, ctx, req); err != nil {
		return maskAny(err)
	}

	c.Config.Logger.Debug(ctx, "controller: executing destroy action, req: %v", req)
	// Destroy.
	if err := c.executeTaskAction(c.Destroy, ctx, req); err != nil {
		return maskAny(err)
	}
	return nil
}

// UpdateOptions represents the options defining the strategy of an update
// process. Lets have a look at how the update process of 3 group slices would
// look like using the given options.
//
//     MaxGrowth    1
//     MinAlive     2
//     ReadySecs    30
//
//     @1 (running)  ->  @1 (stopped/destroyed)
//     @2 (running)  ->  @2 (running)            ->  @2 (stopped/destroyed)
//     @3 (running)  ->  @3 (running)            ->  @3 (running)            ->  @3 (stopped/destroyed)
//                   ->  @1 (submitted/running)  ->  @1 (running)            ->  @1 (running)            ->  @1 (running)
//                                               ->  @2 (submitted/running)  ->  @2 (running)            ->  @2 (running)
//                                                                           ->  @3 (submitted/running)  ->  @3 (running)
//
type UpdateOptions struct {
	// MaxGrowth represents the number of groups allowed to be added at a given
	// time. No more than MaxGrowth groups will be added at the same time during
	// the update process.
	MaxGrowth int

	// MinAlive represents the number of groups required to stay healthy during
	// the update process. No more than MinAlive groups will be removed at the
	// same time during the update process.
	MinAlive int

	// ReadySecs represents the number of seconds required to wait before ending
	// the update process of one group and starting the update process of another
	// group. This is basically a cool down where the update process sleeps
	// before updating the next group.
	ReadySecs int
}

// updateCurrentSliceIDs updates the list of current slice IDs,
// removing the slice that was modified, and adding any new slice IDs.
func (c controller) updateCurrentSliceIDs(ctx context.Context, currentSliceIDs []string, modifiedSliceIDs []string, newSliceIDs []string) []string {
	c.Config.Logger.Debug(ctx, "current slice IDs: %v", currentSliceIDs)
	c.Config.Logger.Debug(ctx, "modified slice IDs: %v", modifiedSliceIDs)
	c.Config.Logger.Debug(ctx, "new slice IDs: %v", newSliceIDs)

	// Remove any slice IDs that have been modified.
	for i, currentSliceID := range currentSliceIDs {
		for _, modifiedSliceID := range modifiedSliceIDs {
			if currentSliceID == modifiedSliceID {
				currentSliceIDs = append(
					currentSliceIDs[:i],
					currentSliceIDs[i+1:]...,
				)
			}
		}
	}

	// And add the new slice IDs.
	currentSliceIDs = append(currentSliceIDs, newSliceIDs...)

	c.Config.Logger.Debug(ctx, "updated slice IDs: %v", currentSliceIDs)

	return currentSliceIDs
}

func (c controller) UpdateWithStrategy(ctx context.Context, req Request, opts UpdateOptions) error {
	c.Config.Logger.Debug(ctx, "controller: running update for group '%v'", req.Group)

	fail := make(chan error, 1)
	numTotal := len(req.SliceIDs)

	done := make(chan struct{}, numTotal)

	var addInProgress, removeInProgress int64

	c.Config.Logger.Debug(ctx, "controller: checking if request is sliceable")
	if !req.isSliceable() {
		return maskAnyf(updateNotAllowedError, "cannot update unsliceable group")
	}

	c.Config.Logger.Debug(ctx, "controller: checking for slice ids in request")
	for _, sliceID := range req.SliceIDs {
		if sliceID == "" {
			return maskAnyf(updateNotAllowedError, "group misses slice ID")
		}
	}

	if numTotal < opts.MinAlive {
		return maskAnyf(updateNotAllowedError, "invalid min alive option")
	}

	// We need to track which slice IDs are currently in use.
	// This list is updated as slices are added and removed.
	currentSliceIDsMutex := sync.Mutex{}
	currentSliceIDs := []string{}
	for _, id := range req.SliceIDs {
		currentSliceIDs = append(currentSliceIDs, id)
	}

	for _, sliceID := range req.SliceIDs {
		newReq := req
		newReq.SliceIDs = []string{sliceID}

		for {
			c.Config.Logger.Debug(ctx, "controller: attempting to add slice: %v", sliceID)

			currentSliceReq := req
			// We try to add slices first. We are only allowed to increase to number
			// slices if the MaxGrowth value minus the additions that are currently
			// in progress is greater than zero.
			// => ((opts.MaxGrowth  - additionInProgress) > 0)
			// See also isGroupAdditionAllowed.
			c.Config.Logger.Debug(
				ctx, "controller: opts.MaxGrowth: %v, numTotal: %v, opts.MinAlive: %v, addInProgress: %v",
				opts.MaxGrowth, numTotal, opts.MinAlive, int(addInProgress),
			)

			c.Config.Logger.Debug(ctx, "controller: currentSliceIDs: %v", currentSliceIDs)
			currentSliceReq.SliceIDs = currentSliceIDs
			ok, err := c.isGroupAdditionAllowed(ctx, currentSliceReq, opts.MaxGrowth, &addInProgress)
			if err != nil {
				return maskAny(err)
			}
			if ok {
				ctx = context.WithValue(ctx, "slice ID", sliceID)
				// we increase the addInProgress counter before starting the goroutine
				// to avoid a race condition in the allowed calculation
				atomic.AddInt64(&addInProgress, 1)
				go func(ctx context.Context) {
					ctx = context.WithValue(ctx, "add slice", sliceID)
					c.Config.Logger.Debug(ctx, "controller: starting to add slice: %v", sliceID)

					newSliceIDs, err := c.addFirst(ctx, newReq, opts)
					if err != nil {
						fail <- maskAny(err)
						return
					}

					currentSliceIDsMutex.Lock()
					currentSliceIDs = c.updateCurrentSliceIDs(ctx, currentSliceIDs, newReq.SliceIDs, newSliceIDs)
					currentSliceIDsMutex.Unlock()

					atomic.AddInt64(&addInProgress, -1)
					done <- struct{}{}
				}(ctx)

				break
			}

			c.Config.Logger.Debug(ctx, "controller: attempting to remove slice: %v", sliceID)
			// remove
			// we are only allowed to remove if the number of minAlive slices
			// minus the ones, that are beeing removed right now is greater than 0
			//=> (minAlive - int(removeInProgress)) > 0)
			c.Config.Logger.Debug(
				ctx, "controller: opts.MinAlive: %v, removeInProgress: %v",
				opts.MinAlive, int(removeInProgress),
			)

			c.Config.Logger.Debug(ctx, "controller: currentSliceIDs: %v", currentSliceIDs)
			currentSliceReq.SliceIDs = currentSliceIDs
			ok, err = c.isGroupRemovalAllowed(ctx, currentSliceReq, opts.MinAlive, &removeInProgress)
			if err != nil {
				return maskAny(err)
			}
			if ok {
				// we increase the removeInProgress counter before starting the goroutine
				// to avoid a race condition in the allowed calculation
				atomic.AddInt64(&removeInProgress, 1)
				go func(ctx context.Context) {
					ctx = context.WithValue(ctx, "remove slice", sliceID)
					c.Config.Logger.Debug(ctx, "controller: starting to remove slice: %v", sliceID)

					newSliceIDs, err := c.removeFirst(ctx, newReq, opts)
					if err != nil {
						fail <- maskAny(err)
						return
					}

					currentSliceIDsMutex.Lock()
					currentSliceIDs = c.updateCurrentSliceIDs(ctx, currentSliceIDs, newReq.SliceIDs, newSliceIDs)
					currentSliceIDsMutex.Unlock()

					atomic.AddInt64(&removeInProgress, -1)
					done <- struct{}{}
				}(ctx)

				break
			}

			time.Sleep(c.WaitSleep)
		}
	}

	tc := 0

	for {
		select {
		case err := <-fail:
			return maskAny(err)
		case <-done:
			tc++
			if tc == numTotal {
				return nil
			}
		case <-time.After(c.WaitTimeout):
			return maskAny(waitTimeoutReachedError)
		}
	}
}
