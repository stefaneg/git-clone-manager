package renderer

import (
	"fmt"
	"time"
	"tools/internal/color"
	"tools/internal/counter"
)

type CloneViewModel struct {
	GroupCount           *counter.Counter
	ProjectCount         *counter.Counter
	CloneCount           *counter.Counter
	ClonedNowCount       *counter.Counter
	ArchivedCloneCounter *counter.Counter
	StartTime            time.Time
}

// Renderer handles rendering counters in different modes
type Renderer struct {
	store *CloneViewModel
	isTTY bool
}

func NewRenderer(store *CloneViewModel, isTTY bool) *Renderer {
	return &Renderer{
		store: store,
		isTTY: isTTY,
	}
}

func (r *Renderer) StartTTYRenderLoop() {
	// Initial placeholder rendering to create space for counters
	r.render()

	for {
		fmt.Printf("\033[%dA", 4)
		r.render()
		time.Sleep(100 * time.Millisecond) // Refresh rate
	}
}

func (r *Renderer) render() {
	fmt.Printf("%s projects in %s groups\n%s git clones (%s archived) \n%s cloned now\n%s seconds\n",
		color.FgMagenta(fmt.Sprintf("%d", r.store.ProjectCount.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.store.GroupCount.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.store.CloneCount.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.store.ArchivedCloneCounter.Count())),
		color.FgMagenta(fmt.Sprintf("%d", r.store.ClonedNowCount.Count())),
		color.FgGreen(fmt.Sprintf("%.2f", time.Since(r.store.StartTime).Seconds())))
}

func (r *Renderer) RenderNonTTY() {
	fmt.Println("Cloning done")
	r.render()
}
