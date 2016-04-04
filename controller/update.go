package controller

import (
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
		ok, err := aggregator.unitHasStatus(groupedStatuses[0], StatusRunning)
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

func (c controller) isGroupRemovalAllowed(ctx context.Context, req Request, minAlive int) (bool, error) {
	c.Config.Logger.Debug(ctx, "controller: checking group removal allowed, req: %v", req)

	numRunning, err := c.getNumRunningSlices(ctx, req)
	if err != nil {
		return false, maskAny(err)
	}

	if numRunning > minAlive {
		c.Config.Logger.Debug(ctx, "controller: group removal allowed")
		return true, nil
	}

	c.Config.Logger.Debug(ctx, "controller: group removal not allowed")
	return false, nil
}

func (c controller) isGroupAdditionAllowed(ctx context.Context, req Request, maxGrowth int) (bool, error) {
	c.Config.Logger.Debug(ctx, "controller: checking group addition allowed, req: %v", req)

	numRunning, err := c.getNumRunningSlices(ctx, req)
	if err != nil {
		return false, maskAny(err)
	}

	if numRunning < maxGrowth {
		c.Config.Logger.Debug(ctx, "controller: group addition allowed")
		return true, nil
	}

	c.Config.Logger.Debug(ctx, "controller: group addition not allowed")
	return false, nil
}

func (c controller) addFirst(ctx context.Context, req Request, opts UpdateOptions) ([]string, error) {
	c.Config.Logger.Debug(ctx, "controller: running addFirst")

	c.Config.Logger.Debug(ctx, "controller: running add worker")
	newReq, err := c.runAddWorker(ctx, req, opts)
	if err != nil {
		return []string{}, maskAny(err)
	}

	c.Config.Logger.Debug(ctx, "controller: checking number of running slices")
	n, err := c.getNumRunningSlices(ctx, newReq)
	if err != nil {
		return []string{}, maskAny(err)
	}
	if n != len(newReq.SliceIDs) {
		return []string{}, maskAnyf(updateFailedError, "addFirst: slice not running: %d != %v", n, newReq.SliceIDs)
	}

	c.Config.Logger.Debug(ctx, "controller: running remove worker")
	err = c.runRemoveWorker(ctx, req)
	if err != nil {
		return []string{}, maskAny(err)
	}

	return newReq.SliceIDs, nil
}

func (c controller) runAddWorker(ctx context.Context, req Request, opts UpdateOptions) (Request, error) {
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
		return []string{}, maskAny(err)
	}

	c.Config.Logger.Debug(ctx, "controller: running add worker")
	newReq, err := c.runAddWorker(ctx, req, opts)
	if err != nil {
		return []string{}, maskAny(err)
	}

	c.Config.Logger.Debug(ctx, "controller: checking number of running slices")
	n, err := c.getNumRunningSlices(ctx, newReq)
	if err != nil {
		return []string{}, maskAny(err)
	}
	if n != len(newReq.SliceIDs) {
		return []string{}, maskAnyf(updateFailedError, "removeFirst: slice not running: %d != %v", n, newReq.SliceIDs)
	}

	return newReq.SliceIDs, nil
}

func (c controller) runRemoveWorker(ctx context.Context, req Request) error {
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
	currentSliceIDs = append(currentSliceIDs, newSliceIDs...)

	c.Config.Logger.Debug(ctx, "updated slice IDs: %v", currentSliceIDs)

	return currentSliceIDs
}

func (c controller) UpdateWithStrategy(ctx context.Context, req Request, opts UpdateOptions) error {
	c.Config.Logger.Debug(ctx, "controller: running update for group '%v'", req.Group)

	fail := make(chan error, 1)
	numTotal := len(req.SliceIDs)
	done := make(chan struct{}, numTotal)
	var addInProgress int64
	var removeInProgress int64

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
	currentSliceIDs := []string{}
	for _, id := range req.SliceIDs {
		currentSliceIDs = append(currentSliceIDs, id)
	}

	for _, sliceID := range req.SliceIDs {
		newReq := req
		newReq.SliceIDs = []string{sliceID}

		// Track the number of times we attempted to make a change, and failed.
		numFailedChangeAttempts := 0

		for {
			currentSliceReq := req

			changeAttemptMade := false

			c.Config.Logger.Debug(ctx, "controller: attempting to add slice: %v", sliceID)
			// add
			maxGrowth := opts.MaxGrowth + numTotal - opts.MinAlive - int(addInProgress)

			currentSliceReq.SliceIDs = currentSliceIDs
			ok, err := c.isGroupAdditionAllowed(ctx, currentSliceReq, maxGrowth)
			if err != nil {
				return maskAny(err)
			}
			if ok {
				ctx = context.WithValue(ctx, "add slice", sliceID)
				c.Config.Logger.Debug(ctx, "controller: starting to add slice: %v", sliceID)
				atomic.AddInt64(&addInProgress, 1)

				newSliceIDs, err := c.addFirst(ctx, newReq, opts)
				if err != nil {
					fail <- maskAny(err)
					return maskAny(err)
				}

				currentSliceIDs = c.updateCurrentSliceIDs(ctx, currentSliceIDs, newReq.SliceIDs, newSliceIDs)

				atomic.AddInt64(&addInProgress, -1)
				done <- struct{}{}

				changeAttemptMade = true
				break
			}

			c.Config.Logger.Debug(ctx, "controller: attempting to remove slice: %v", sliceID)
			// remove
			minAlive := opts.MinAlive + int(removeInProgress)

			currentSliceReq.SliceIDs = currentSliceIDs
			ok, err = c.isGroupRemovalAllowed(ctx, currentSliceReq, minAlive)
			if err != nil {
				return maskAny(err)
			}
			if ok {
				ctx = context.WithValue(ctx, "remove slice", sliceID)
				c.Config.Logger.Debug(ctx, "controller: starting to remove slice: %v", sliceID)
				atomic.AddInt64(&removeInProgress, 1)

				newSliceIDs, err := c.removeFirst(ctx, newReq, opts)
				if err != nil {
					fail <- maskAny(err)
					return maskAny(err)
				}

				currentSliceIDs = c.updateCurrentSliceIDs(ctx, currentSliceIDs, newReq.SliceIDs, newSliceIDs)

				atomic.AddInt64(&removeInProgress, -1)
				done <- struct{}{}

				changeAttemptMade = true
				break
			}

			c.Config.Logger.Debug(ctx, "controller: finished attempts to make changes")
			if !changeAttemptMade {
				c.Config.Logger.Warning(ctx, "controller: failed to make any changes")
				numFailedChangeAttempts++
			}
			// If we've failed too many times, just give up completely :(
			if numFailedChangeAttempts > c.Config.MaxFailedChangeAttempts {
				return maskAnyf(updateFailedError, "reached max failed change attempts limit")
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
