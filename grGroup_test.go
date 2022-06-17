package grgroup_test

import (
	"errors"
	"fmt"
	"github.com/jiz12/grgroup"
	"testing"
	"time"
)

func BenchmarkGrGroup(b *testing.B) {

	ch := make(chan string, 10)
	res := make([]string, 0)
	g, err := grgroup.NewGrGroup(10)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		i := i
		msg := fmt.Sprintf("发生了错误%d", i)
		g.Go(func() error {
			ch <- test(i)
			return errors.New(msg)
		})
	}

	go func() {
		for i := range ch {
			res = append(res, i)
		}
		fmt.Println("我在结束了呢！！")
	}()

	err = g.Wait()
	close(ch)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}

func test(i int) string {
	time.Sleep(time.Second * time.Duration(i))
	msg := fmt.Sprintf("我在执行了呢！%d", i)
	fmt.Println(msg)
	return msg
}
