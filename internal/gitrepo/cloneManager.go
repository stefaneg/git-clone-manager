package gitrepo

import (
	"fmt"
	"gcm/internal/counter"
	"gcm/internal/log"
	"sync"
)

func CloneRepositories(repositories <-chan *Repository, cloneCounter *counter.Counter, errorChannel chan error) {
	cloneWaitGroup := sync.WaitGroup{}
	for {
		receivedRepo, ok := <-repositories
		if !ok {
			break
		}
		cloneWaitGroup.Add(1)
		go func() {
			defer cloneWaitGroup.Done()
			err := receivedRepo.Clone()
			if err != nil {
				errorChannel <- fmt.Errorf("failed to clone project %s: %v", receivedRepo.Name, err)
				return
			}
			cloneCounter.Add(1)
		}()
	}
	cloneWaitGroup.Wait()
}

func FilterCloneNeeded(checkCloneChannel <-chan *Repository, archivedCounter *counter.Counter, clonedCounter *counter.Counter) chan *Repository {
	gitCloneChannel := make(chan *Repository, 20)
	checkWaitGroup := sync.WaitGroup{}
	go func() {
		for {
			receivedRepo, ok := <-checkCloneChannel
			if !ok {
				logger.Log.Tracef("%s \n", "Clone channel close, wait for last clone to finish, then breaking")
				break
			}
			if receivedRepo.Archived && receivedRepo.CloneOptions.CloneArchived() {
				archivedCounter.Add(1)
			}

			needsCloning, _ := receivedRepo.CheckNeedsCloning()
			cloned, _ := receivedRepo.IsCloned()
			// TODO ADD ERROR HANDLING CHANNEL
			if cloned || needsCloning {
				clonedCounter.Add(1)
			}

			if needsCloning {
				checkWaitGroup.Add(1)
				go func() {
					defer checkWaitGroup.Done()
					logger.Log.Debugf("Adding %s to clone queue ", receivedRepo.Name)
					gitCloneChannel <- receivedRepo
				}()
			} else {
				logger.Log.Tracef("%s \n", "Clone not needed")
			}
		}
		checkWaitGroup.Wait()
		close(gitCloneChannel)
	}()
	return gitCloneChannel
}
