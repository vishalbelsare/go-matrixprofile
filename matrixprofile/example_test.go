package matrixprofile

import (
	"fmt"
	"os"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

func Points(a []float64, n int) plotter.XYs {
	pts := make(plotter.XYs, n)
	for i := 0; i < n; i++ {
		pts[i].X = float64(i)
		if i < len(a) {
			pts[i].Y = a[i]
		}
	}
	return pts
}

func CreatePlot(pts []plotter.XYs, labels []string) (*plot.Plot, error) {
	if labels != nil && len(pts) != len(labels) {
		return nil, fmt.Errorf("number of XYs, %d, does not match number of labels, %d", len(pts), len(labels))
	}

	p, err := plot.New()
	if err != nil {
		return p, err
	}

	for i := 0; i < len(pts); i++ {
		if labels == nil {
			err = plotutil.AddLines(p, "", pts[i])
		} else {
			err = plotutil.AddLines(p, labels[i], pts[i])
		}
		if err != nil {
			return p, err
		}
	}
	return p, err
}

func PlotMP(sigPts, mpPts, cacPts plotter.XYs, motifPts [][]plotter.XYs, filename string) error {
	var err error
	rows, cols := 3, 2
	plots := make([][]*plot.Plot, rows)

	plots[0] = make([]*plot.Plot, cols)
	plots[1] = make([]*plot.Plot, cols)
	plots[2] = make([]*plot.Plot, cols)

	plots[0][0], err = CreatePlot([]plotter.XYs{sigPts}, []string{"data"})
	if err != nil {
		return err
	}

	plots[1][0], err = CreatePlot([]plotter.XYs{mpPts}, []string{"matrix profile"})
	if err != nil {
		return err
	}

	plots[2][0], err = CreatePlot([]plotter.XYs{cacPts}, []string{"cac"})
	if err != nil {
		return err
	}

	plots[0][1], err = CreatePlot(motifPts[0], nil)
	if err != nil {
		return err
	}

	plots[1][1], err = CreatePlot(motifPts[1], nil)
	if err != nil {
		return err
	}

	plots[2][1], err = CreatePlot(motifPts[2], nil)
	if err != nil {
		return err
	}

	img := vgimg.New(vg.Points(1200), vg.Points(600))
	dc := draw.New(img)

	t := draw.Tiles{
		Rows: rows,
		Cols: cols,
	}

	canvases := plot.Align(plots, t, dc)
	for j := 0; j < rows; j++ {
		for i := 0; i < cols; i++ {
			if plots[j][i] != nil {
				plots[j][i].Draw(canvases[j][i])
			}
		}
	}

	w, err := os.Create(filename)
	if err != nil {
		return err
	}

	png := vgimg.PngCanvas{Canvas: img}
	_, err = png.WriteTo(w)
	return err
}

func Example() {
	sin := generateSin(1, 5, 0, 0, 100, 2)
	sin2 := generateSin(0.25, 10, 0, 0.75, 100, 1)
	sig := append(sin, sin2...)
	noise := generateNoise(0.3, len(sin2)*2)
	sig = append(sig, noise...)
	sig = append(sig, sin2...)
	sig = append(sig, noise...)
	noise = generateNoise(0.1, len(sig))
	sig = sigAdd(sig, noise)

	var m, k int
	var r float64
	m = 32
	k = 3
	r = 3
	mp, err := New(sig, nil, m)
	if err != nil {
		panic(err)
	}

	if err = mp.Stmp(); err != nil {
		panic(err)
	}

	_, _, cac := mp.Segment()

	motifs, err := mp.TopKMotifs(k, r)
	if err != nil {
		panic(err)
	}

	sigPts := Points(sig, len(sig))
	mpPts := Points(mp.MP, len(sig))
	cacPts := Points(cac, len(sig))
	motifPts := make([][]plotter.XYs, k)

	for i := 0; i < k; i++ {
		motifPts[i] = make([]plotter.XYs, len(motifs[i].Idx))
	}

	for i := 0; i < k; i++ {
		for j, idx := range motifs[i].Idx {
			motifPts[i][j] = Points(sig[idx:idx+m], m)
		}
	}

	if err = PlotMP(sigPts, mpPts, cacPts, motifPts, "mp_sine.png"); err != nil {
		panic(err)
	}

	fmt.Println("Saved png file result to mp_sine.png")
	// Output: Saved png file result to mp_sine.png
}

func ExampleMatrixProfile_Stmp() {
	// generate a signal mainly composed of sine waves and switches
	// frequencies, amplitude, and offset midway through

	// amplitude of 1, frequency of 5Hz, sampling frequency of 100 Hz,
	// time of 2 seconds
	sin := generateSin(1, 5, 0, 0, 100, 2)

	// amplitude of 0.25, frequency of 10Hz, offset of 0.75, sampling
	// frequency of 100 Hz, time of 1 second
	sin2 := generateSin(0.25, 10, 0, 0.75, 100, 1)
	sig := append(sin, sin2...)

	// noise with an amplitude of 0.1
	noise := generateNoise(0.1, len(sig))
	sig = sigAdd(sig, noise)

	// create a new MatrixProfile struct using the signal and a
	// subsequence length of 32. The second subsequence is set to nil
	// so we perform a self join.
	mp, err := New(sig, nil, 32)
	if err != nil {
		panic(err)
	}

	// run the STMP algorithm with self join. The matrix profile
	// will be stored in mp.MP and the matrix profile index will
	// be stored in mp.Idx
	if err = mp.Stmp(); err != nil {
		panic(err)
	}
}

func ExampleMatrixProfile_Stamp() {
	// generate a signal mainly composed of sine waves and switches
	// frequencies, amplitude, and offset midway through

	// amplitude of 1, frequency of 5Hz, sampling frequency of 100 Hz,
	// time of 2 seconds
	sin := generateSin(1, 5, 0, 0, 100, 2)

	// amplitude of 0.25, frequency of 10Hz, offset of 0.75, sampling
	// frequency of 100 Hz, time of 1 second
	sin2 := generateSin(0.25, 10, 0, 0.75, 100, 1)
	sig := append(sin, sin2...)

	// noise with an amplitude of 0.1
	noise := generateNoise(0.1, len(sig))
	sig = sigAdd(sig, noise)

	// create a new MatrixProfile struct using the signal and a
	// subsequence length of 32. The second subsequence is set to nil
	// so we perform a self join.
	mp, err := New(sig, nil, 32)
	if err != nil {
		panic(err)
	}

	// run the STAMP algorithm with self join and a sample of 0.2 of
	// all subsequences. The matrix profile will be stored in mp.MP
	// and the matrix profile index will be stored in mp.Idx
	if err = mp.Stamp(0.2); err != nil {
		panic(err)
	}

}

func ExampleMatrixProfile_Segment() {
	// generate a signal mainly composed of sine waves and switches
	// frequencies, amplitude, and offset midway through

	// amplitude of 1, frequency of 5Hz, sampling frequency of 100 Hz,
	// time of 2 seconds
	sin := generateSin(1, 5, 0, 0, 100, 2)

	// amplitude of 0.25, frequency of 10Hz, offset of 0.75, sampling
	// frequency of 100 Hz, time of 1 second
	sin2 := generateSin(0.25, 10, 0, 0.75, 100, 1)
	sig := append(sin, sin2...)

	// noise with an amplitude of 0.1
	noise := generateNoise(0.01, len(sig))
	sig = sigAdd(sig, noise)

	// create a new MatrixProfile struct using the signal and a
	// subsequence length of 32. The second subsequence is set to nil
	// so we perform a self join.
	mp, err := New(sig, nil, 32)
	if err != nil {
		panic(err)
	}

	// run the STMP algorithm with self join. The matrix profile
	// will be stored in mp.MP and the matrix profile index will
	// be stored in mp.Idx
	if err = mp.Stmp(); err != nil {
		panic(err)
	}

	// segment the timeseries using the number of arc crossings over
	// each index in the matrix profile index
	idx, cac, _ := mp.Segment()
	fmt.Printf("Signal change foud at index: %d\n", idx)
	fmt.Printf("Corrected Arc Curve (CAC) value: %.3f\n", cac)

	// Output:
	// Signal change foud at index: 194
	// Corrected Arc Curve (CAC) value: 0.000
}

func ExampleMatrixProfile_TopKMotifs() {
	// generate a signal mainly composed of sine waves and switches
	// frequencies, amplitude, and offset midway through

	// amplitude of 1, frequency of 5Hz, sampling frequency of 100 Hz,
	// time of 2 seconds
	sin := generateSin(1, 5, 0, 0, 100, 2)

	// amplitude of 0.25, frequency of 10Hz, offset of 0.75, sampling
	// frequency of 100 Hz, time of 1 second
	sin2 := generateSin(0.25, 10, 0, 0.75, 100, 1)
	sig := append(sin, sin2...)

	// noise with an amplitude of 0.1
	noise := generateNoise(0.01, len(sig))
	sig = sigAdd(sig, noise)

	// create a new MatrixProfile struct using the signal and a
	// subsequence length of 32. The second subsequence is set to nil
	// so we perform a self join.
	mp, err := New(sig, nil, 32)
	if err != nil {
		panic(err)
	}

	// run the STMP algorithm with self join. The matrix profile
	// will be stored in mp.MP and the matrix profile index will
	// be stored in mp.Idx
	if err = mp.Stmp(); err != nil {
		panic(err)
	}

	// finds the top 3 motifs in the signal. Motif groups include
	// all subsequences that are within 2 times the distance of the
	// original motif pair
	motifs, err := mp.TopKMotifs(2, 2)
	if err != nil {
		panic(err)
	}

	for i, mg := range motifs {
		fmt.Printf("Motif Group %d\n", i)
		fmt.Printf("  %d motifs\n", len(mg.Idx))
		fmt.Printf("  minimum distance of %.3f\n", mg.MinDist)
	}

	// Output:
	// Motif Group 0
	//   9 motifs
	//   minimum distance of 0.021
	// Motif Group 1
	//   7 motifs
	//   minimum distance of 0.090
}