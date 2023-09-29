package graphs

import (
	"bytes"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"weight-tracker/db"
)

func Graph(data []db.Weight) ([]byte, error) {
	svg, err := plotData(data)
	if err != nil {
		return nil, errors.Wrap(err, "plot data")
	}

	return svg, nil
}

func plotData(data []db.Weight) ([]byte, error) {
	p := plot.New()

	p.Title.Text = "Weight" // TODO: allow change
	p.X.Label.Text = "t"
	p.Y.Label.Text = "Weight"

	var pts plotter.XYs

	xticks := plot.TimeTicks{Format: "2006-01-02\n15:04"}

	for _, w := range data {
		pts = append(pts, plotter.XY{
			X: float64(w.T.Unix()),
			Y: w.Weight,
		})
	}

	p.X.Tick.Marker = xticks

	err := plotutil.AddLinePoints(p,
		"weight (kg)", pts,
	)
	if err != nil {
		return nil, errors.Wrap(err, "add line points")
	}

	w, err := p.WriterTo(8*vg.Inch, 4*vg.Inch, "svg")
	buf := bytes.NewBuffer(nil)
	_, err = w.WriteTo(buf)
	if err != nil {
		return nil, errors.Wrap(err, "write to")
	}

	return buf.Bytes(), nil
}
