package streamer

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

type StreamBandwithCalculator interface {
	// NewStream called for each new stream
	NewStream(streamname string, size uint64)
	// Add adds amount of bytes read from network
	Add(n uint64)
	// Finish is called when stream is closed
	Finish()
}

type BandwithCalc struct {
	n          uint64 //Bytes
	start      time.Time
	lastprint  time.Time
	duration   time.Duration
	size       uint64
	w          io.Writer
	streamname string
}

func NewBandwithCalc(w io.Writer) *BandwithCalc {
	b := &BandwithCalc{
		start:     time.Now(),
		lastprint: time.Time{},
		w:         w,
	}

	return b
}

func (b *BandwithCalc) calcIn(munit int64) float64 {
	duration := b.duration.Seconds()
	if duration < 1 {
		duration = 1
	}

	bandwidth := float64(b.n) / float64(munit) / float64(duration)
	return bandwidth
}

func (b *BandwithCalc) total(munit int64) float64 {
	bandwidth := float64(b.n) / float64(munit)
	return bandwidth
}

func (b *BandwithCalc) percentage() int64 {
	if b.n == 0 {
		return 0
	}
	bandwidth := float64(b.n) / float64(b.size) * 100
	return int64(bandwidth)
}

func (b *BandwithCalc) printOnSecond() {
	since := time.Since(b.lastprint)
	if since.Seconds() > 1 { //Printing only if there are changes
		b.print()
		b.lastprint = time.Now()
	}
}

func (b *BandwithCalc) print() {
	speed := b.calcIn(MB)
	total := b.total(MB)
	percentage := b.percentage()
	s := fmt.Sprintf("\033[999D%s     %d%% %.2fMB %.2f MB/s\033[K", b.streamname, percentage, total, speed)
	fmt.Fprint(b.w, s)
}

func (b *BandwithCalc) NewStream(streamname string, n uint64) {
	b.start = time.Now()
	b.lastprint = time.Time{}
	b.streamname = streamname
	b.size = n
}

//When adding bytes we calculate duration, to be more precise
func (b *BandwithCalc) Add(n uint64) {
	b.n += n
	b.duration = time.Since(b.start)
	b.printOnSecond()
}

func (b *BandwithCalc) Finish() {
	b.print()
}
