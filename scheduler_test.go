package profile

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func newTestScheduler() (*Scheduler, *[]time.Time) {
	times := []time.Time{}
	s := newScheduler(func() {
		times = append(times, time.Now())
	})

	return s, &times
}

func assertTimeWithin(t *testing.T, t1, t2 time.Time, delta time.Duration) {
	sub := t2.Sub(t1)
	if sub < 0 {
		sub = -sub
	}

	if sub > delta {
		t.Errorf("Expected %d to be within %d ns of %d, but got %d",
			t1.UnixNano(), delta, t2.UnixNano(), sub)
	}
}

func TestScheduleAtOnce(t *testing.T) {
	s, times := newTestScheduler()
	start := time.Now()
	s.At(start.Add(100 * time.Millisecond))

	go s.start()
	defer s.end()

	time.Sleep(200 * time.Millisecond)
	assertTimeWithin(t, (*times)[0], start.Add(100*time.Millisecond), time.Millisecond*20)
	assert.Equal(t, 1, len(*times))
}

func TestScheduleAfterMany(t *testing.T) {
	s, times := newTestScheduler()
	start := time.Now()
	s.After(500 * time.Millisecond).
		Every(80 * time.Millisecond).
		For(200 * time.Millisecond)

	go s.start()
	defer s.end()

	time.Sleep(800 * time.Millisecond)
	assertTimeWithin(t, (*times)[0], start.Add(500*time.Millisecond), time.Millisecond*20)
	assertTimeWithin(t, (*times)[1], start.Add(580*time.Millisecond), time.Millisecond*20)
	assertTimeWithin(t, (*times)[2], start.Add(660*time.Millisecond), time.Millisecond*20)
	assert.Equal(t, 3, len(*times))
}

func TestScheduleImmediatelyUntil(t *testing.T) {
	s, times := newTestScheduler()
	start := time.Now()
	s.Every(80 * time.Millisecond).Until(start.Add(300 * time.Millisecond))

	go s.start()
	defer s.end()

	time.Sleep(400 * time.Millisecond)
	assertTimeWithin(t, (*times)[0], start.Add(0*time.Millisecond), time.Millisecond*20)
	assertTimeWithin(t, (*times)[1], start.Add(80*time.Millisecond), time.Millisecond*20)
	assertTimeWithin(t, (*times)[2], start.Add(160*time.Millisecond), time.Millisecond*20)
	assertTimeWithin(t, (*times)[3], start.Add(240*time.Millisecond), time.Millisecond*20)
	assert.Equal(t, 4, len(*times))
}
