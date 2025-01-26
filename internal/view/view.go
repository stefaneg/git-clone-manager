package view

type View interface {
	Render(width int) (lines int)
}
