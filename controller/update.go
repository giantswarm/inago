package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/giantswarm/inago/task"
)

type updateJobDesc struct {
	Request Request
}

func (c controller) getNumRunningSlices(req Request) (int, error) {
	unitStatusList, err := c.groupStatusWithValidate(req)
	if err != nil {
		return 0, maskAny(err)
	}

	var numRunning int
	for _, us := range unitStatusList {
		ok, err := unitHasStatus(us, StatusRunning)
		if err != nil {
			return 0, maskAny(err)
		}
		if ok {
			numRunning++
		}
	}

	return numRunning, nil
}

func (c controller) createJobQueues(req Request, opts UpdateOptions) (chan updateJobDesc, chan updateJobDesc, error) {
	addQueue := make(chan updateJobDesc, len(req.SliceIDs))
	removeQueue := make(chan updateJobDesc, len(req.SliceIDs))

	numRunning, err := c.getNumRunningSlices(req)
	if err != nil {
		return nil, nil, maskAny(err)
	}
	numAllowedToRemove := numRunning - opts.MinAlive

	fmt.Printf("numRunning: %#v\n", numRunning)
	fmt.Printf("opts.MinAlive: %#v\n", opts.MinAlive)
	fmt.Printf("numAllowedToRemove: %#v\n", numAllowedToRemove)

	var numRemovalQueued int
	for _, sliceID := range req.SliceIDs {
		jobDesc := updateJobDesc{
			Request: Request{
				Group:    req.Group,
				SliceIDs: []string{sliceID},
				Units:    req.Units,
			},
		}

		addQueue <- jobDesc
		fmt.Printf("job added to addQueue\n")

		if numAllowedToRemove >= numRemovalQueued {
			removeQueue <- jobDesc
			numRemovalQueued++
			fmt.Printf("job added to removeQueue\n")
		}
	}

	return addQueue, removeQueue, nil
}

func (c controller) isGroupRemovalAllowed(req Request, opts UpdateOptions) (bool, error) {
	numRunning, err := c.getNumRunningSlices(req)
	if err != nil {
		return false, maskAny(err)
	}

	if numRunning > opts.MinAlive {
		return true, nil
	}

	return false, nil
}

func (c controller) isGroupAdditionAllowed(req Request, numTotal int, opts UpdateOptions) (bool, error) {
	numRunning, err := c.getNumRunningSlices(req)
	if err != nil {
		return false, maskAny(err)
	}

	if numRunning < numTotal+opts.MaxGrowth {
		return true, nil
	}

	return false, nil
}

func (c controller) runAddWorker(jobDesc updateJobDesc, fail chan<- error, readySecs int) {
	fmt.Printf("call runAddWorker\n")

	// Submit.
	taskObject, err := c.Submit(jobDesc.Request)
	if err != nil {
		fail <- maskAny(err)
		return
	}
	closer := make(<-chan struct{})
	taskObject, err = c.WaitForTask(taskObject.ID, closer)
	if err != nil {
		fail <- maskAny(err)
		return
	}
	if task.HasFailedStatus(taskObject) {
		fail <- maskAny(fmt.Errorf(taskObject.Error))
		return
	}

	// Start.
	taskObject, err = c.Start(jobDesc.Request)
	if err != nil {
		fail <- maskAny(err)
		return
	}
	closer = make(<-chan struct{})
	taskObject, err = c.WaitForTask(taskObject.ID, closer)
	if err != nil {
		fail <- maskAny(err)
		return
	}
	if task.HasFailedStatus(taskObject) {
		fail <- maskAny(fmt.Errorf(taskObject.Error))
		return
	}

	fmt.Printf("runAddWorker sleep %ds\n", readySecs)
	time.Sleep(time.Duration(readySecs) * time.Second)

	fmt.Printf("end runAddWorker\n")
}

func (c controller) runRemoveWorker(jobDesc updateJobDesc, fail chan<- error) {
	fmt.Printf("call runRemoveWorker\n")

	// Stop.
	taskObject, err := c.Stop(jobDesc.Request)
	if err != nil {
		fail <- maskAny(err)
		return
	}
	closer := make(<-chan struct{})
	taskObject, err = c.WaitForTask(taskObject.ID, closer)
	if err != nil {
		fail <- maskAny(err)
		return
	}
	if task.HasFailedStatus(taskObject) {
		fail <- maskAny(fmt.Errorf(taskObject.Error))
		return
	}

	// Destroy.
	taskObject, err = c.Destroy(jobDesc.Request)
	if err != nil {
		fail <- maskAny(err)
		return
	}
	closer = make(<-chan struct{})
	taskObject, err = c.WaitForTask(taskObject.ID, closer)
	if err != nil {
		fail <- maskAny(err)
		return
	}
	if task.HasFailedStatus(taskObject) {
		fail <- maskAny(fmt.Errorf(taskObject.Error))
		return
	}
	fmt.Printf("end runRemoveWorker\n")
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
	done := make(chan struct{}, 1)
	fail := make(chan error, 1)
	addQueue := make(chan updateJobDesc, len(req.SliceIDs))
	addQueue, removeQueue, err := c.createJobQueues(req, opts)
	if err != nil {
		return maskAny(err)
	}
	numTotal := len(req.SliceIDs)
	fmt.Printf("numTotal: %#v\n", numTotal)

	m := sync.Mutex{}
	var alreadyUpdated int

	for i := 0; i < opts.MaxGrowth; i++ {
		go func() {
			for jobDesc := range addQueue {
				for {
					ok, err := c.isGroupAdditionAllowed(req, numTotal, opts)
					if err != nil {
						fail <- maskAny(err)
						return
					}
					if !ok {
						time.Sleep(c.WaitSleep)
					}
					break
				}

				c.runAddWorker(jobDesc, fail, opts.ReadySecs)
				fmt.Printf("add jobDesc to removeQueue\n")
				removeQueue <- jobDesc
			}
			fmt.Printf("finished addQueue\n")
		}()
	}

	for {
		select {
		case jobDesc := <-removeQueue:
			go func(jobDesc updateJobDesc) {
				for {
					ok, err := c.isGroupRemovalAllowed(req, opts)
					if err != nil {
						fail <- maskAny(err)
						return
					}
					if !ok {
						time.Sleep(c.WaitSleep)
					}
					break
				}

				c.runRemoveWorker(jobDesc, fail)
				m.Lock()
				alreadyUpdated++
				if alreadyUpdated == numTotal {
					done <- struct{}{}
				}
				m.Unlock()
			}(jobDesc)
		case err := <-fail:
			return maskAny(err)
		case <-done:
			return nil
		case <-time.After(c.WaitTimeout):
			return maskAny(waitTimeoutReachedError)
		}
	}
}
