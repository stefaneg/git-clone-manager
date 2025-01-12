package view

type CompositeView struct {
	headers []View
	views   []View
	footers []View
}

func NewCompositeView(views []View) *CompositeView {
	return &CompositeView{views: views}
}

func (cv *CompositeView) AddHeader(view View) {
	cv.headers = append(cv.headers, view)
}

func (cv *CompositeView) AddView(view View) {
	cv.views = append(cv.views, view)
}

func (cv *CompositeView) AddFooter(view View) {
	cv.footers = append(cv.footers, view)
}

func (cv *CompositeView) Render(w int) int {
	totalLines := 0
	for _, view := range cv.headers {
		lines := view.Render(w)
		totalLines += lines
	}
	for _, view := range cv.views {
		lines := view.Render(w)
		totalLines += lines
	}
	for _, view := range cv.footers {
		lines := view.Render(w)
		totalLines += lines
	}
	return totalLines
}
