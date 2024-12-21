package gitrepo

import (
	"sync"
	"tools/internal/color"
	"tools/internal/log"
)

func CloneRepositories(repositories <-chan *Repository) int {
	cloneWaitGroup := sync.WaitGroup{}
	cloneCount := 0
	for {
		receivedRepo, ok := <-repositories
		if !ok {
			break
		}
		cloneCount++
		cloneWaitGroup.Add(1)
		go func() {
			defer cloneWaitGroup.Done()
			err := receivedRepo.Clone()
			if err != nil {
				logger.Log.Errorf("Failed to clone project %s: %v", color.FgRed(receivedRepo.Name), err)
			}
		}()
	}
	cloneWaitGroup.Wait()
	return cloneCount
}

func FilterCloneNeeded(checkCloneChannel <-chan *Repository) chan *Repository {
	gitCloneChannel := make(chan *Repository, 20)
	checkWaitGroup := sync.WaitGroup{}
	go func() {
		for {
			receivedRepo, ok := <-checkCloneChannel
			if !ok {
				logger.Log.Tracef("%s \n", "Clone channel close, wait for last clone to finish, then breaking")
				break
			}
			needsCloning, _ := receivedRepo.CheckNeedsCloning()
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
