package graphs

import (
	"bytes"
	"fmt"
	"github.com/go-gorp/gorp/v3"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"weight-tracker/db"
)

func Graph(dbmap *gorp.DbMap) ([]byte, error) {
	var data []db.Weight
	if _, err := dbmap.Select(&data, "select * from weight order by t"); err != nil {
		return nil, errors.Wrap(err, "select")
	}

	for _, w := range data {
		fmt.Println(w)
	}

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
