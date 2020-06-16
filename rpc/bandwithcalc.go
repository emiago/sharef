package rpc

import (
	"fmt"
	"io"
	"time"
)

const (
	KB int64 = 1024
	MB int64 = 1024 * 1024
	GB int64 = 1024 * 1024 * 1024
)

type BandwithCalc struct {
	n         uint64 //Bytes
	start     time.Time
	lastprint time.Time
	duration  time.Duration
	size      uint64
}

func NewBandwithCalc(size uint64) *BandwithCalc {
	b := &BandwithCalc{
		start:     time.Now(),
		lastprint: time.Time{},
		size:      size,
	}

	return b
}

func (b *BandwithCalc) ResetTimer() {
	b.start = time.Now()
}

//When adding bytes we calculate duration, to be more precise
func (b *BandwithCalc) Add(n uint64) {
	b.n += n
	b.duration = time.Since(b.start)
}

func (b *BandwithCalc) CalcIn(munit int64) float64 {
	duration := b.duration.Seconds()
	if duration < 1 {
		duration = 1
	}

	bandwidth := float64(b.n) / float64(munit) / float64(duration)
	return bandwidth
}

func (b *BandwithCalc) Total(munit int64) float64 {
	bandwidth := float64(b.n) / float64(munit)
	return bandwidth
}

func (b *BandwithCalc) Percentage() int64 {
	bandwidth := float64(b.n) / float64(b.size) * 100
	return int64(bandwidth)
}

func (b *BandwithCalc) FprintOnSecond(w io.Writer, streamname string) {
	since := time.Since(b.lastprint)
	if since.Seconds() > 1 { //Printing only if there are changes
		// speed := b.CalcIn(MB)
		// total := b.Total(MB)
		// percentage := b.Percentage()
		// s := fmt.Sprintf("\033[999D%s     %d%% %.2fMB %.2f MB/s\033[K", streamname, percentage, total, speed)
		// fmt.Fprint(w, s)
		fmt.Fprint(w, b.Sprint(streamname))
		b.lastprint = time.Now()
	}
}

func (b *BandwithCalc) Sprint(streamname string) string {
	speed := b.CalcIn(MB)
	total := b.Total(MB)
	percentage := b.Percentage()
	s := fmt.Sprintf("\033[999D%s     %d%% %.2fMB %.2f MB/s\033[K", streamname, percentage, total, speed)
	return s
}

func (b *BandwithCalc) String() string {
	res := fmt.Sprintf("%.2f KB/s", b.CalcIn(KB))
	return res
}
