// Copyright 2019 by Donald Wilson. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"
	
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
)

// widgets holds the widgets used by this demo.
type widgets struct {
	segDist  *segmentdisplay.SegmentDisplay
	input    *textinput.TextInput
	rollT    *text.Text
	spGreen  *sparkline.SparkLine
	spRed    *sparkline.SparkLine
	gauge    *gauge.Gauge
	heartLC  *linechart.LineChart
	barChart *barchart.BarChart
	// donut    *donut.Donut
	leftB  *button.Button
	rightB *button.Button
	sineLC *linechart.LineChart

	buttons *layoutButtons
}

// newWidgets sets up the widgets.
func newWidgets(ctx context.Context, c *container.Container) (*widgets, error) {
	updateText := make(chan string)
	sd, err := newSegmentDisplay(ctx, updateText)
	if err != nil {
		return nil, err
	}

	input, err := newTextInput(updateText)
	if err != nil {
		return nil, err
	}

	rollT, err := newRollText(ctx)
	if err != nil {
		return nil, err
	}
	spGreen, spRed, err := newSparkLines(ctx)
	if err != nil {
		return nil, err
	}
	g, err := newGauge(ctx)
	if err != nil {
		return nil, err
	}

	heartLC, err := newHeartbeat(ctx)
	if err != nil {
		return nil, err
	}

	bc, err := newBarChart(ctx)
	if err != nil {
		return nil, err
	}

	// don, err := newDonut(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	leftB, rightB, sineLC, err := newSines(ctx)
	if err != nil {
		return nil, err
	}

	return &widgets{
		segDist:  sd,
		input:    input,
		rollT:    rollT,
		spGreen:  spGreen,
		spRed:    spRed,
		gauge:    g,
		heartLC:  heartLC,
		barChart: bc,
		// donut:    don,
		leftB:  leftB,
		rightB: rightB,
		sineLC: sineLC,
	}, nil
}

// newBarChart returns a BarcChart that displays random values on multiple bars.
func newBarChart(ctx context.Context) (*barchart.BarChart, error) {
	bc, err := barchart.New(
		barchart.BarColors([]cell.Color{
			cell.ColorNumber(33),
			cell.ColorNumber(39),
			cell.ColorNumber(45),
			cell.ColorNumber(51),
			cell.ColorNumber(81),
			cell.ColorNumber(87),
		}),
		barchart.ValueColors([]cell.Color{
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
		}),
		barchart.ShowValues(),
	)
	if err != nil {
		return nil, err
	}

	const (
		bars = 6
		max  = 100
	)
	values := make([]int, bars)
	go periodic(ctx, 1*time.Second, func() error {
		for i := range values {
			values[i] = int(rand.Int31n(max + 1))
		}

		return bc.Values(values, max)
	})
	return bc, nil
}

// distance is a thread-safe int value used by the newSince method.
// Buttons write it and the line chart reads it.
type distance struct {
	v  int
	mu sync.Mutex
}

// add adds the provided value to the one stored.
func (d *distance) add(v int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.v += v
}

// get returns the current value.
func (d *distance) get() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.v
}

// newSines returns a line chart that displays multiple sine series and two buttons.
// The left button shifts the second series relative to the first series to
// the left and the right button shifts it to the right.
func newSines(ctx context.Context) (left, right *button.Button, lc *linechart.LineChart, err error) {
	var inputs []float64
	for i := 0; i < 200; i++ {
		v := math.Sin(float64(i) / 100 * math.Pi)
		inputs = append(inputs, v)
	}

	sineLc, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	step1 := 0
	secondDist := &distance{v: 100}
	go periodic(ctx, prototypes.redrawInterval/3, func() error {
		step1 = (step1 + 1) % len(inputs)
		if err := lc.Series("first", rotateFloats(inputs, step1),
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorBlue)),
		); err != nil {
			return err
		}

		step2 := (step1 + secondDist.get()) % len(inputs)
		return lc.Series("second", rotateFloats(inputs, step2), linechart.SeriesCellOpts(cell.FgColor(cell.ColorWhite)))
	})

	// diff is the difference a single button press adds or removes to the
	// second series.
	const diff = 20
	leftB, err := button.New("(l)eft", func() error {
		secondDist.add(diff)
		return nil
	},
		button.GlobalKey('l'),
		button.WidthFor("(r)ight"),
		button.FillColor(cell.ColorNumber(220)),
	)
	if err != nil {
		return nil, nil, nil, err
	}

	rightB, err := button.New("(r)ight", func() error {
		secondDist.add(-diff)
		return nil
	},
		button.GlobalKey('r'),
		button.FillColor(cell.ColorNumber(196)),
	)
	if err != nil {
		return nil, nil, nil, err
	}
	return leftB, rightB, sineLc, nil
}

// setLayout sets the specified layout.
func setLayout(c *container.Container, w *widgets, lt layoutType) error {
	gridOpts, err := gridLayout(w, lt)
	if err != nil {
		return err
	}
	return c.Update(prototypes.rootID, gridOpts...)
}

// layoutButtons are buttons that change the layout.
type layoutButtons struct {
	allB  *button.Button
	textB *button.Button
	spB   *button.Button
	lcB   *button.Button
}

// newLayoutButtons returns buttons that dynamically switch the layouts.
func newLayoutButtons(c *container.Container, w *widgets) (*layoutButtons, error) {
	opts := []button.Option{
		button.WidthFor("sparklines"),
		button.FillColor(cell.ColorNumber(220)),
		button.Height(1),
	}

	allB, err := button.New("all", func() error {
		return setLayout(c, w, layoutAll)
	}, opts...)
	if err != nil {
		return nil, err
	}

	textB, err := button.New("text", func() error {
		return setLayout(c, w, layoutText)
	}, opts...)
	if err != nil {
		return nil, err
	}

	spB, err := button.New("sparklines", func() error {
		return setLayout(c, w, layoutSparkLines)
	}, opts...)
	if err != nil {
		return nil, err
	}

	lcB, err := button.New("linechart", func() error {
		return setLayout(c, w, layoutLineChart)
	}, opts...)
	if err != nil {
		return nil, err
	}

	return &layoutButtons{
		allB:  allB,
		textB: textB,
		spB:   spB,
		lcB:   lcB,
	}, nil
}
