package hashcash

import (
	"time"

	"github.com/bullgare/pow-ddos-protection/internal/usecase/contracts"
)

const DifficultyChangeStep = 3

const tickDuration = 5 * time.Second

func NewDifficultyManager(
	targetRPS float64,
	levelChangeStep int,
) (_ *DifficultyManager, stopFunc func()) {
	manager := &DifficultyManager{
		targetRPS: targetRPS,
		step:      levelChangeStep,

		chIncrRequests: make(chan struct{}),
		chGetRPS:       make(chan chan float64),
		chGetCurLevel:  make(chan chan int),
		chSetCurLevel:  make(chan int),

		chQuit: make(chan struct{}),
	}

	curLevel := 30 // we don't want to start too fast, so it's not 0
	curReqBucket := int64(0)
	targetReqsPerBucket := int64(tickDuration.Seconds() * targetRPS)
	prevReqBucket := targetReqsPerBucket // by default, we expect we are hitting the target exactly, so it's not 0

	manager.startLoop(curLevel, curReqBucket, prevReqBucket)
	stopFn := func() {
		close(manager.chQuit)
	}
	return manager, stopFn
}

var _ contracts.DifficultyManager = &DifficultyManager{}

type DifficultyManager struct {
	targetRPS float64
	step      int

	chIncrRequests chan struct{}
	chGetRPS       chan chan float64
	chGetCurLevel  chan chan int
	chSetCurLevel  chan int

	chQuit chan struct{}
}

func (m *DifficultyManager) IncrRequests() {
	go func() {
		m.chIncrRequests <- struct{}{}
	}()
}

func (m *DifficultyManager) GetDifficulty() int {
	curRPS := m.getRPS()
	curLevel := m.getCurrentLevel()

	if curRPS > m.targetRPS {
		curLevel += m.step
	} else if curRPS < m.targetRPS {
		curLevel -= m.step
	}

	curLevel = max(0, curLevel)
	curLevel = min(curLevel, 100)

	go func() {
		m.chSetCurLevel <- curLevel
	}()

	return curLevel
}

func (m *DifficultyManager) getRPS() float64 {
	chSendRPS := make(chan float64)
	go func() {
		m.chGetRPS <- chSendRPS
	}()

	return <-chSendRPS
}

func (m *DifficultyManager) getCurrentLevel() int {
	chSendLevel := make(chan int)
	go func() {
		m.chGetCurLevel <- chSendLevel
	}()

	return <-chSendLevel
}

// we only manipulate curLevel, curReqBucket, prevReqBucket fields here.
// yes, it could be a bit inaccurate potentially as we are not using mutexes, but it is faster, which is more important here.
func (m *DifficultyManager) startLoop(
	curLevel int,
	curReqBucket int64,
	prevReqBucket int64,
) {
	go func(
		curLevel int,
		curReqBucket int64,
		prevReqBucket int64,
	) {
		ticker := time.NewTicker(tickDuration)
		defer ticker.Stop()

		start := time.Now()
		for {
			// here we synchronously do 1 of 6 things:
			// increment number of requests, change buckets, send average RPS for the tick period, get or set current level, or quit.
			select {
			case <-m.chIncrRequests:
				curReqBucket++
			case <-ticker.C:
				prevReqBucket = curReqBucket
				curReqBucket = 0
				start = time.Now()
			case chSendRPS := <-m.chGetRPS:
				timeElapsed := time.Since(start)
				chSendRPS <- m.calculateRPS(curReqBucket, prevReqBucket, timeElapsed, tickDuration)
			case chSendLevel := <-m.chGetCurLevel:
				chSendLevel <- curLevel
			case level := <-m.chSetCurLevel:
				curLevel = level
			case <-m.chQuit:
				return
			}
		}
	}(curLevel, curReqBucket, prevReqBucket)
}

// FIXME add unit-tests.
func (m *DifficultyManager) calculateRPS(curReqBucket, prevReqBucket int64, timeElapsed, tickDuration time.Duration) float64 {
	if timeElapsed > tickDuration {
		timeElapsed = tickDuration
	}

	fractionFromPrevBucket := float64(tickDuration.Nanoseconds()-timeElapsed.Nanoseconds()) / float64(tickDuration.Nanoseconds())

	totalRequests := float64(curReqBucket) + (float64(prevReqBucket) * fractionFromPrevBucket)

	return totalRequests / tickDuration.Seconds()
}

var _ contracts.DifficultyManager = NoOpDifficultyManagerForClient{}

type NoOpDifficultyManagerForClient struct{}

func (m NoOpDifficultyManagerForClient) IncrRequests() {}

func (m NoOpDifficultyManagerForClient) GetDifficulty() int { return 0 }
