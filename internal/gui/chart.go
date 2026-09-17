package gui

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// maxChartSamples caps how many points the live chart keeps, matching the
// TUI's asciigraph window so both frontends behave the same way.
const maxChartSamples = 120

// LineChart is a minimal line chart for live throughput display. It avoids
// pulling in a charting library by drawing directly with canvas primitives,
// reusing a fixed pool of Line objects so redraws don't churn allocations.
type LineChart struct {
	widget.BaseWidget

	mu      sync.Mutex
	samples []float64
}

func NewLineChart() *LineChart {
	c := &LineChart{}
	c.ExtendBaseWidget(c)
	return c
}

// SetSamples replaces the chart's data and triggers a redraw.
func (c *LineChart) SetSamples(samples []float64) {
	c.mu.Lock()
	if len(samples) > maxChartSamples {
		samples = samples[len(samples)-maxChartSamples:]
	}
	c.samples = append([]float64(nil), samples...)
	c.mu.Unlock()
	c.Refresh()
}

func (c *LineChart) snapshot() []float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.samples
}

func (c *LineChart) CreateRenderer() fyne.WidgetRenderer {
	r := &lineChartRenderer{chart: c}
	r.bg = canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	r.maxLabel = canvas.NewText("", theme.Color(theme.ColorNameForegroundOnPrimary))
	r.maxLabel.TextSize = 11
	r.zeroLine = canvas.NewLine(theme.Color(theme.ColorNameSeparator))
	for i := 0; i < maxChartSamples; i++ {
		l := canvas.NewLine(theme.Color(theme.ColorNamePrimary))
		l.StrokeWidth = 2
		l.Hide()
		r.segments = append(r.segments, l)
	}
	return r
}

type lineChartRenderer struct {
	chart    *LineChart
	bg       *canvas.Rectangle
	zeroLine *canvas.Line
	maxLabel *canvas.Text
	segments []*canvas.Line
}

func (r *lineChartRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.rebuild(size)
}

func (r *lineChartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 140)
}

func (r *lineChartRenderer) Refresh() {
	r.rebuild(r.chart.Size())
	r.bg.FillColor = theme.Color(theme.ColorNameInputBackground)
	canvas.Refresh(r.chart)
}

func (r *lineChartRenderer) Objects() []fyne.CanvasObject {
	objs := []fyne.CanvasObject{r.bg, r.zeroLine, r.maxLabel}
	for _, l := range r.segments {
		objs = append(objs, l)
	}
	return objs
}

func (r *lineChartRenderer) Destroy() {}

func (r *lineChartRenderer) rebuild(size fyne.Size) {
	samples := r.chart.snapshot()

	const padTop = 18.0
	const padBottom = 4.0
	plotH := float64(size.Height) - padTop - padBottom
	if plotH < 1 {
		plotH = 1
	}

	r.zeroLine.Position1 = fyne.NewPos(0, float32(padTop+plotH))
	r.zeroLine.Position2 = fyne.NewPos(size.Width, float32(padTop+plotH))

	if len(samples) < 2 {
		for _, l := range r.segments {
			l.Hide()
		}
		r.maxLabel.Text = ""
		return
	}

	maxV := samples[0]
	for _, v := range samples {
		if v > maxV {
			maxV = v
		}
	}
	if maxV <= 0 {
		maxV = 1
	}
	r.maxLabel.Text = fmt.Sprintf("%.0f Mbps", maxV)
	r.maxLabel.Move(fyne.NewPos(4, 0))

	n := len(samples)
	stepX := float64(size.Width) / float64(n-1)
	mapY := func(v float64) float32 {
		return float32(padTop + plotH - (v/maxV)*plotH)
	}

	for i := 0; i < maxChartSamples; i++ {
		if i >= n-1 {
			r.segments[i].Hide()
			continue
		}
		x1 := float32(float64(i) * stepX)
		x2 := float32(float64(i+1) * stepX)
		r.segments[i].Position1 = fyne.NewPos(x1, mapY(samples[i]))
		r.segments[i].Position2 = fyne.NewPos(x2, mapY(samples[i+1]))
		r.segments[i].Show()
	}
}
