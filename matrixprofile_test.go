package matrixprofile

import (
	"math"
	"os"
	"testing"

	"github.com/matrix-profile-foundation/go-matrixprofile/av"
)

func TestNew(t *testing.T) {
	testdata := []struct {
		a           []float64
		b           []float64
		m           int
		expectedErr bool
	}{
		{[]float64{}, []float64{}, 2, true},
		{[]float64{1, 1, 1, 1, 1}, []float64{}, 2, true},
		{[]float64{1, 1, 1, 1, 1}, nil, 2, false},
		{[]float64{1, 1, 1, 1, 1}, nil, 6, true},
		{[]float64{1, 1}, []float64{1, 1, 1, 1, 1, 1, 1, 1}, 3, true},
		{[]float64{}, []float64{1, 1, 1, 1, 1}, 2, true},
		{[]float64{1, 2, 3, 4, 5}, []float64{1, 1, 1, 1, 1}, 2, false},
		{[]float64{1, 2, 3, 4, 5}, []float64{1, 1, 1, 1, 1}, 1, true},
		{[]float64{1, 2, 3, 4, 5}, []float64{1, 1, 1, 1, 1}, 4, false},
	}

	for _, d := range testdata {
		_, err := New(d.a, d.b, d.m)
		if d.expectedErr && err == nil {
			t.Errorf("Expected an error, but got none for %v", d)
			return
		}
		if !d.expectedErr && err != nil {
			t.Errorf("Expected no error, but got %v for %v", err, d)
			return
		}
	}
}

func TestApplyAV(t *testing.T) {
	mprof := []float64{4, 6, 10, 2, 1, 0, 1, 2, 0, 0, 1, 2, 6}

	testdata := []struct {
		b          []float64
		m          int
		av         av.AV
		expectedMP []float64
	}{
		{[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, 4, av.Default, mprof},
	}

	var mp MatrixProfile
	var err error
	var out []float64
	for _, d := range testdata {
		newMP := make([]float64, len(mprof))
		copy(newMP, mprof)
		mp = MatrixProfile{B: d.b, M: d.m, MP: newMP, AV: d.av}
		out, err = mp.ApplyAV()
		if err != nil {
			t.Fatal(err)
		}

		if len(out) != len(d.expectedMP) {
			t.Errorf("Expected %d elements, but got %d, %+v", len(d.expectedMP), len(out), d)
			break
		}
		for i := 0; i < len(out); i++ {
			if math.Abs(float64(out[i]-d.expectedMP[i])) > 1e-7 {
				t.Errorf("Expected %v,\nbut got\n%v for %+v", d.expectedMP, out, d)
				break
			}
		}
	}
}

func TestSave(t *testing.T) {
	ts := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	m := 3
	p, err := New(ts, nil, m)
	p.Compute(NewComputeOpts())
	filepath := "./mp.json"
	err = p.Save(filepath, "json")
	if err != nil {
		t.Errorf("Received error while saving matrix profile, %v", err)
	}
	if err = os.Remove(filepath); err != nil {
		t.Errorf("Could not remove file, %s, %v", filepath, err)
	}
}

func TestLoad(t *testing.T) {
	ts := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	m := 3
	p, err := New(ts, nil, m)
	p.Compute(NewComputeOpts())
	filepath := "./mp.json"
	if err = p.Save(filepath, "json"); err != nil {
		t.Errorf("Received error while saving matrix profile, %v", err)
	}

	newP := &MatrixProfile{}
	if err = newP.Load(filepath, "json"); err != nil {
		t.Errorf("Failed to load %s, %v", filepath, err)
	}

	if err = os.Remove(filepath); err != nil {
		t.Errorf("Could not remove file, %s, %v", filepath, err)
	}

	if newP.M != m {
		t.Errorf("Expected m of %d, but got %d", m, newP.M)
	}
	if len(newP.A) != len(ts) {
		t.Errorf("Expected timeseries length of %d, but got %d", len(ts), len(newP.A))
	}

}

func TestMPDist(t *testing.T) {
	testData := []struct {
		a        []float64
		b        []float64
		m        int
		expected float64
	}{
		{
			[]float64{1, 2, 3, 4, 3, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			[]float64{0, 0, 0, 0, 0, 0, 0, 0, -1, -2, -3, -2, -1, 0, 1, 2, 1, 0},
			5,
			0,
		},
		{
			[]float64{1, 2, 3, 4, 3, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			[]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 1, 0},
			5,
			0,
		},
	}
	for _, d := range testData {
		res, err := MPDist(d.a, d.b, d.m, nil)
		if err != nil {
			t.Errorf("Did not expect to get an error, %v", err)
		}
		if math.Abs(res-d.expected) > 1e-4 {
			t.Errorf("Expected %.6f, but got %.6f", d.expected, res)
		}
	}
}
