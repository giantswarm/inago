package controller

import (
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"

	"github.com/giantswarm/inago/task"
)

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
		return maskAny(fmt.Errorf(taskObject.Error))
	}
	return nil
}

func (c controller) getNumRunningSlices(req Request) (int, error) {
	groupStatus, err := c.groupStatus(req)
	if IsUnitNotFound(err) {
		return 0, nil
	} else if err != nil {
		return 0, maskAny(err)
	}

	grouped, err := UnitStatusList(groupStatus).Group()
	if err != nil {
		return 0, maskAny(err)
	}

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

	return len(sliceIDs), nil
}

func (c controller) isGroupRemovalAllowed(req Request, minAlive int) (bool, error) {
	numRunning, err := c.getNumRunningSlices(req)
	if err != nil {
		return false, maskAny(err)
	}

	if numRunning > minAlive {
		return true, nil
	}

	return false, nil
}

func (c controller) isGroupAdditionAllowed(req Request, maxGrowth int) (bool, error) {
	numRunning, err := c.getNumRunningSlices(req)
	if err != nil {
		return false, maskAny(err)
	}

	if numRunning < maxGrowth {
		return true, nil
	}

	return false, nil
}

func (c controller) addFirst(ctx context.Context, req Request, opts UpdateOptions) error {
	newReq, err := c.runAddWorker(ctx, req, opts)
	if err != nil {
		return maskAny(err)
	}
	n, err := c.getNumRunningSlices(newReq)
	if err != nil {
		return maskAny(err)
	}
	if n != len(newReq.SliceIDs) {
		return maskAnyf(updateFailedError, "slice not running: %v", newReq.SliceIDs)
	}
	err = c.runRemoveWorker(ctx, req)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (c controller) runAddWorker(ctx context.Context, req Request, opts UpdateOptions) (Request, error) {
	// Create new random IDs.
	newReq, err := c.ExtendWithRandomSliceIDs(req)
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

func (c controller) removeFirst(ctx context.Context, req Request, opts UpdateOptions) error {
	err := c.runRemoveWorker(ctx, req)
	if err != nil {
		return maskAny(err)
	}
	newReq, err := c.runAddWorker(ctx, req, opts)
	if err != nil {
		return maskAny(err)
	}
	n, err := c.getNumRunningSlices(newReq)
	if err != nil {
		return maskAny(err)
	}
	if n != len(newReq.SliceIDs) {
		return maskAnyf(updateFailedError, "slice not running: %v", newReq.SliceIDs)
	}

	return nil
}

func (c controller) runRemoveWorker(ctx context.Context, req Request) error {
	// Stop.
	if err := c.executeTaskAction(c.Stop, ctx, req); err != nil {
		return maskAny(err)
	}

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

func (c controller) UpdateWithStrategy(ctx context.Context, req Request, opts UpdateOptions) error {
	fail := make(chan error, 1)
	numTotal := len(req.SliceIDs)
	done := make(chan struct{}, numTotal)
	var addInProgress int64
	var removeInProgress int64

	for _, sliceID := range req.SliceIDs {
		if sliceID == "" {
			return maskAnyf(updateNotAllowedError, "group misses slice ID")
		}
	}

	if numTotal < opts.MinAlive {
		return maskAnyf(updateNotAllowedError, "invalid min alive option")
	}

	for _, sliceID := range req.SliceIDs {
		newReq := req
		newReq.SliceIDs = []string{sliceID}

		for {
			// add
			maxGrowth := opts.MaxGrowth + numTotal - int(addInProgress)
			ok, err := c.isGroupAdditionAllowed(req, maxGrowth)
			if err != nil {
				return maskAny(err)
			}
			if ok {
				go func() {
					atomic.AddInt64(&addInProgress, 1)
					err := c.addFirst(ctx, newReq, opts)
					if err != nil {
						fail <- maskAny(err)
						return
					}
					atomic.AddInt64(&addInProgress, -1)
					done <- struct{}{}
				}()

				break
			}

			// remove
			minAlive := opts.MinAlive + int(removeInProgress)
			ok, err = c.isGroupRemovalAllowed(req, minAlive)
			if err != nil {
				return maskAny(err)
			}
			if ok {
				go func() {
					atomic.AddInt64(&removeInProgress, 1)
					err := c.removeFirst(ctx, newReq, opts)
					if err != nil {
						fail <- maskAny(err)
						return
					}
					atomic.AddInt64(&removeInProgress, -1)
					done <- struct{}{}
				}()

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
