package rpc

import (
	"math"
	"math/rand"
	"os"
	"testing"
	"time"

	"gotest.tools/assert"
)

func bandwithCalcFakeData(b *BandwithCalc, n uint64, t time.Duration) {
	b.Add(n)
	b.duration = t
}

func TestBandWithCalc(t *testing.T) {
	b := NewBandwithCalc(os.Stdout)
	b.NewStream("", 4096)

	bandwithCalcFakeData(b, 4096, 1*time.Second)
	//Lets calculate
	res := b.calcIn(KB)
	assert.Equal(t, math.Round(res), math.Round(4.00))
}

func TestBandWithLoopCalc(t *testing.T) {
	b := NewBandwithCalc(os.Stdout)
	b.NewStream("", 0)

	var totalbytes uint64
	for i := 1; i <= 10; i++ {
		//Make some random bytes between to 1 to 64K
		nbytes := uint64(1 + rand.Intn(int(64*KB)))
		totalbytes += nbytes
		duration := time.Duration(i) * time.Second

		bandwithCalcFakeData(b, nbytes, duration)

		res := b.calcIn(KB)
		expected := float64(totalbytes) / float64(KB) / float64(duration.Seconds())
		assert.Equal(t, math.Round(res), math.Round(expected))
	}
}

func TestBandWithPercentage(t *testing.T) {
	var third uint64 = 51234
	var totalbytes uint64 = 3 * third

	b := NewBandwithCalc(os.Stdout)
	b.NewStream("", totalbytes)

	b.Add(third)
	res := b.percentage()
	assert.Equal(t, res, int64(33), "Not 33%")

	b.Add(third)
	res = b.percentage()
	assert.Equal(t, res, int64(66), "Not 66%")

	b.Add(third)
	res = b.percentage()
	assert.Equal(t, res, int64(100), "Not 100%")
}
