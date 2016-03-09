package controller

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/giantswarm/inago/task"
)

func (c controller) getNumRunningSlices(req Request) (int, error) {
	usl, err := c.groupStatus(req)
	if IsUnitNotFound(err) {
		return 0, nil
	} else if err != nil {
		return 0, maskAny(err)
	}

	var sliceIDs []string
	for _, us := range usl {
		if contains(sliceIDs, us.Slice) {
			// We already tracked this ID. Ho ahead.
			continue
		}
		ok, err := unitHasStatus(us, StatusRunning)
		if err != nil {
			return 0, maskAny(err)
		}
		if !ok {
			continue
		}
		sliceIDs = append(sliceIDs, us.Slice)
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

func (c controller) addFirst(req Request, opts UpdateOptions) error {
	req, err := c.runAddWorker(req, opts)
	if err != nil {
		return maskAny(err)
	}
	err = c.runRemoveWorker(req)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (c controller) runAddWorker(req Request, opts UpdateOptions) (Request, error) {
	newReq := req
	oldReq := req

	// Create new random IDs.
	var err error
	newReq, err = c.ExtendWithRandomSliceIDs(newReq)
	if err != nil {
		return Request{}, maskAny(err)
	}

	// Submit.
	taskObject, err := c.Submit(newReq)
	if err != nil {
		return Request{}, maskAny(err)
	}
	closer := make(<-chan struct{})
	taskObject, err = c.WaitForTask(taskObject.ID, closer)
	if err != nil {
		return Request{}, maskAny(err)
	}
	if task.HasFailedStatus(taskObject) {
		return Request{}, maskAny(fmt.Errorf(taskObject.Error))
	}

	// Start.
	taskObject, err = c.Start(newReq)
	if err != nil {
		return Request{}, maskAny(err)
	}
	closer = make(<-chan struct{})
	taskObject, err = c.WaitForTask(taskObject.ID, closer)
	if err != nil {
		return Request{}, maskAny(err)
	}
	if task.HasFailedStatus(taskObject) {
		return Request{}, maskAny(err)
	}

	time.Sleep(time.Duration(opts.ReadySecs) * time.Second)

	return oldReq, nil
}

func (c controller) removeFirst(req Request, opts UpdateOptions) error {
	err := c.runRemoveWorker(req)
	if err != nil {
		return maskAny(err)
	}
	_, err = c.runAddWorker(req, opts)
	if err != nil {
		return maskAny(err)
	}

	return nil
}

func (c controller) runRemoveWorker(req Request) error {
	// Stop.
	taskObject, err := c.Stop(req)
	if err != nil {
		return maskAny(err)
	}
	closer := make(<-chan struct{})
	taskObject, err = c.WaitForTask(taskObject.ID, closer)
	if err != nil {
		return maskAny(err)
	}
	if task.HasFailedStatus(taskObject) {
		return maskAny(err)
	}

	// Destroy.
	taskObject, err = c.Destroy(req)
	if err != nil {
		return maskAny(err)
	}
	closer = make(<-chan struct{})
	taskObject, err = c.WaitForTask(taskObject.ID, closer)
	if err != nil {
		return maskAny(err)
	}
	if task.HasFailedStatus(taskObject) {
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

func (c controller) UpdateWithStrategy(req Request, opts UpdateOptions) error {
	fail := make(chan error, 1)
	numTotal := len(req.SliceIDs)
	done := make(chan struct{}, numTotal)
	var addInProgress int64
	var removeInProgress int64

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
					v := atomic.AddInt64(&addInProgress, 1)
					addInProgress = v
					err := c.addFirst(newReq, opts)
					if err != nil {
						fail <- maskAny(err)
						return
					}
					v = atomic.AddInt64(&addInProgress, -1)
					addInProgress = v
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
					v := atomic.AddInt64(&removeInProgress, 1)
					removeInProgress = v
					err := c.removeFirst(newReq, opts)
					if err != nil {
						fail <- maskAny(err)
						return
					}
					v = atomic.AddInt64(&removeInProgress, -1)
					removeInProgress = v
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
