// Package util is a set of utility functions that are used throughout the matrixprofile package.
package util

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/stat"
)

// ZNormalize computes a z-normalized version of a slice of floats.
// This is represented by y[i] = (x[i] - mean(x))/std(x)
func ZNormalize(ts []float64) ([]float64, error) {
	var i int

	if len(ts) == 0 {
		return nil, fmt.Errorf("slice does not have any data")
	}

	m := stat.Mean(ts, nil)

	out := make([]float64, len(ts))
	for i = 0; i < len(ts); i++ {
		out[i] = ts[i] - m
	}

	var std float64
	for _, val := range out {
		std += val * val
	}
	std = math.Sqrt(std / float64(len(out)))

	if std == 0 {
		return out, fmt.Errorf("standard deviation is zero")
	}

	for i = 0; i < len(ts); i++ {
		out[i] = out[i] / std
	}

	return out, nil
}

// MovMeanStd computes the mean and standard deviation of each sliding
// window of m over a slice of floats. This is done by one pass through
// the data and keeping track of the cumulative sum and cumulative sum
// squared.  s between these at intervals of m provide a total of O(n)
// calculations for the standard deviation of each window of size m for
// the time series ts.
func MovMeanStd(ts []float64, m int) ([]float64, []float64, error) {
	if m <= 1 {
		return nil, nil, fmt.Errorf("length of slice must be greater than 1")
	}

	if m > len(ts) {
		return nil, nil, fmt.Errorf("m cannot be greater than length of slice")
	}

	var i int

	c := make([]float64, len(ts)+1)
	csqr := make([]float64, len(ts)+1)
	for i = 0; i < len(ts)+1; i++ {
		if i == 0 {
			c[i] = 0
			csqr[i] = 0
		} else {
			c[i] = ts[i-1] + c[i-1]
			csqr[i] = ts[i-1]*ts[i-1] + csqr[i-1]
		}
	}

	mean := make([]float64, len(ts)-m+1)
	std := make([]float64, len(ts)-m+1)
	for i = 0; i < len(ts)-m+1; i++ {
		mean[i] = (c[i+m] - c[i]) / float64(m)
		std[i] = math.Sqrt((csqr[i+m]-csqr[i])/float64(m) - mean[i]*mean[i])
	}

	return mean, std, nil
}

// ApplyExclusionZone performs an in place operation on a given matrix
// profile setting distances around an index to +Inf
func ApplyExclusionZone(profile []float64, idx, zoneSize int) {
	startIdx := 0
	if idx-zoneSize > startIdx {
		startIdx = idx - zoneSize
	}
	endIdx := len(profile)
	if idx+zoneSize < endIdx {
		endIdx = idx + zoneSize
	}
	for i := startIdx; i < endIdx; i++ {
		profile[i] = math.Inf(1)
	}
}

func MuInvN(a []float64, w int) ([]float64, []float64) {
	mu := Sum2s(a, w)
	sig := make([]float64, len(a)-w+1)
	h := make([]float64, len(a))
	r := make([]float64, len(a))

	var mu_a, c float64
	var a1, a2, a3, p, s, x, z float64
	bigNum := math.Pow(2.0, 27.0) + 1
	for i := 0; i < len(mu); i++ {
		for j := i; j < i+w; j++ {
			mu_a = a[j] - mu[i]
			h[j] = mu_a * mu_a

			c = bigNum * mu_a
			a1 = c - (c - mu_a)
			a2 = mu_a - a1
			a3 = a1 * a2
			r[j] = a2*a2 - (((h[j] - a1*a1) - a3) - a3)
		}

		p = h[i]
		s = r[i]

		for j := i + 1; j < i+w; j++ {
			x = p + h[j]
			z = x - p
			s += ((p - (x - z)) + (h[j] - z)) + r[j]
			p = x
		}

		if p+s == 0 {
			sig[i] = 0
		} else {
			sig[i] = 1 / math.Sqrt(p+s)
		}
	}
	return mu, sig
}

func Sum2s(a []float64, w int) []float64 {
	if len(a) < w {
		return nil
	}
	p := a[0]
	s := 0.0
	var x, z float64
	for i := 1; i < w; i++ {
		x = p + a[i]
		z = x - p
		s += (p - (x - z)) + (a[i] - z)
		p = x
	}

	res := make([]float64, len(a)-w+1)
	res[0] = (p + s) / float64(w)
	for i := w; i < len(a); i++ {
		x = p - a[i-w]
		z = x - p
		s += (p - (x - z)) - (a[i-w] + z)
		p = x

		x = p + a[i]
		z = x - p
		s += (p - (x - z)) + (a[i] - z)
		p = x

		res[i-w+1] = (p + s) / float64(w)
	}

	return res
}

func BinarySplit(lb, ub int) []int {
	if ub < lb {
		return []int{}
	}
	res := make([]int, 1, ub-lb+1)
	res[0] = lb
	if ub == lb {
		return res
	}

	ranges := []*idxRange{&idxRange{lb + 1, ub}}

	var r *idxRange
	var mid int
	for {
		if len(ranges) == 0 {
			break
		}
		// pop first element
		r = ranges[0]
		copy(ranges, ranges[1:])
		ranges = ranges[:len(ranges)-1]

		mid = (r.upper + r.lower) / 2
		res = append(res, mid)

		if r.upper < r.lower {
			continue
		}

		l, r := split(r.lower, r.upper, mid)
		if l != nil {
			ranges = append(ranges, l)
		}
		if r != nil {
			ranges = append(ranges, r)
		}
	}
	return res
}

type idxRange struct {
	lower int
	upper int
}

func split(lower, upper, mid int) (*idxRange, *idxRange) {
	var l *idxRange
	var r *idxRange

	if lower < upper {
		if mid-1 >= lower {
			l = &idxRange{lower, mid - 1}
		}
		if upper >= mid+1 {
			r = &idxRange{mid + 1, upper}
		}
	}

	return l, r
}

// Batch indicates which index to start at and how many to process from that
// index.
type Batch struct {
	Idx  int
	Size int
}

// DiagBatchingScheme computes a more balanced batching scheme based on the
// diagonal nature of computing matrix profiles. Later batches get more to
// work on since those operate on less data in the matrix.
func DiagBatchingScheme(l, p int) []Batch {
	numElem := float64(l*(l+1)) / float64(2*p)
	batchScheme := make([]Batch, p)
	var pi, sum int
	for i := 0; i < l+1; i++ {
		sum += i
		batchScheme[p-pi-1].Size += 1
		if float64(sum) > numElem {
			sum = 0
			pi += 1
		}
	}

	for i := 1; i < p; i++ {
		batchScheme[i].Idx = batchScheme[i-1].Idx + batchScheme[i-1].Size
	}

	return batchScheme
}

// P2E converts a slice of pearson correlation values to euclidean distances. This
// is only valid for z-normalized time series.
func P2E(mp []float64, w int) {
	for i := 0; i < len(mp); i++ {
		// caps pearson correlation to 1 in case there are floating point accumulated errors
		if mp[i] > 1 {
			mp[i] = 1
		}
		mp[i] = math.Sqrt(2 * float64(w) * (1 - mp[i]))
	}
}

// E2P converts a slice of euclidean distances to pearson correlation values. This
// is only valid for z-normalized time series. Negative pearson correlation values will not be
// discovered
func E2P(mp []float64, w int) {
	for i := 0; i < len(mp); i++ {
		mp[i] = 1 - mp[i]*mp[i]/(2*float64(w))
		// caps pearson correlation to 1 in case there are floating point accumulated errors
		if mp[i] > 1 {
			mp[i] = 1
		}
		if mp[i] < 0 {
			mp[i] = 0
		}
	}
}
