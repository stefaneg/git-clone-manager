package gitrepo

import (
	"fmt"
	"gcm/internal/counter"
	"gcm/internal/log"
	"sync"
)

func CloneRepositories(repositories <-chan GitRepo, cloneCounter *counter.Counter, errorChannel chan error) {
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
				errorChannel <- fmt.Errorf("failed to clone project %s: %v", receivedRepo.GetName(), err)
				return
			}
			cloneCounter.Add(1)
		}()
	}
	cloneWaitGroup.Wait()
}

func FilterCloneNeeded(
	repositories <-chan GitRepo,
	archivedCounter *counter.Counter,
	clonedCounter *counter.Counter,
	errorChan chan error,
) chan GitRepo {
	gitCloneChannel := make(chan GitRepo, 20)
	checkWaitGroup := sync.WaitGroup{}
	go func() {
		for {
			receivedRepo, ok := <-repositories
			if !ok {
				logger.Log.Tracef("%s \n", "Clone errorChan close, wait for last clone to finish, then breaking")
				break
			}
			if receivedRepo.IsArchived() && receivedRepo.GetCloneOptions().CloneArchived() {
				archivedCounter.Add(1)
			}

			needsCloning, err := receivedRepo.CheckNeedsCloning()
			if err != nil {
				errorChan <- fmt.Errorf("error checking if project needs cloning %s: %v", receivedRepo.GetName(), err)
				continue
			}

			cloned, err := receivedRepo.IsCloned()
			if err != nil {
				errorChan <- fmt.Errorf("error checking clone status %s: %v", receivedRepo.GetName(), err)
				continue
			}
			if cloned || needsCloning {
				clonedCounter.Add(1)
			}

			if needsCloning {
				checkWaitGroup.Add(1)
				go func() {
					defer checkWaitGroup.Done()
					logger.Log.Debugf("Adding %s to clone queue ", receivedRepo.GetName())
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
