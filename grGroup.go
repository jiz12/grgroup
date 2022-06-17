package grgroup

import (
	"errors"
	"golang.org/x/sync/singleflight"
	"strconv"
	"strings"
	"sync"
)

const (
	exceedMaxGrNumError string = "maxGrNum must be greater than 0"
)

type grGroup struct {
	mu       sync.Mutex
	wg       sync.WaitGroup
	sf       singleflight.Group
	count    int32
	isClose  bool
	maxGrNum int32
	limit    chan struct{}
	errors   []error
}

func NewGrGroup(maxGrNum int32) (*grGroup, error) {
	if maxGrNum < 1 {
		return nil, errors.New(exceedMaxGrNumError)
	}
	group := &grGroup{
		maxGrNum: maxGrNum,
		limit:    make(chan struct{}, maxGrNum),
	}
	return group, nil
}

func (g *grGroup) Go(fn func() error) {
	if g.isClose {
		g.sf.Do(strconv.Itoa(int(g.count)), func() (interface{}, error) {
			g.restart()
			return nil, nil
		})
	}
	g.wg.Add(1)
	g.limit <- struct{}{}
	go func(fn func() error) {
		defer func() {
			<-g.limit
			g.wg.Done()
		}()
		err := fn()
		if err != nil {
			g.mu.Lock()
			g.errors = append(g.errors, err)
			g.mu.Unlock()
		}
	}(fn)
}

func (g *grGroup) restart() {
	g.isClose = false
	g.errors = make([]error, 0)
}

func (g *grGroup) Wait() error {
	g.wg.Wait()
	g.isClose = true
	g.count++
	return g.error()
}

func (g *grGroup) error() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.errors) == 0 {
		return nil
	}
	errMsg := make([]string, len(g.errors))
	for i, err := range g.errors {
		errMsg[i] = err.Error()
	}
	return errors.New(strings.Join(errMsg, ";"))
}
