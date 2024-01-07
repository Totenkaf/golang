package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"sync"
)

var globalCache = make(map[uint64]string)

func RunPipeline(cmds ...cmd) {
	wg := &sync.WaitGroup{}

	inCh := make(chan interface{})

	for cmdIndex, cmdFunc := range cmds {
		wg.Add(1)
		outCh := make(chan interface{})
		go func(index int, task cmd, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			task(in, out)
		}(cmdIndex, cmdFunc, inCh, outCh)
		inCh = outCh
	}
	wg.Wait()
}

func selectUserWorker(
	data string,
	out chan<- interface{},
	wg *sync.WaitGroup,
	mu *sync.Mutex,
) {
	defer wg.Done()
	user := GetUser(data)
	if _, exist := globalCache[user.ID]; !exist {
		mu.Lock()
		globalCache[user.ID] = user.Email
		mu.Unlock()
		out <- user
	}
}

func SelectUsers(in, out chan interface{}) {
	// 	in - string
	// 	out - User

	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for val := range in {
		var data string

		switch val.(type) {
		case string:
			data = val.(string)
		default:
			fmt.Printf("Not string type. Ignore")
			continue
		}
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go selectUserWorker(data, out, wg, mu)
		}
	}
	wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User
	// 	out - MsgID
	wg := &sync.WaitGroup{}
	selectQuotaCh := make(chan struct{}, 2)

	for val := range in {
		var data User
		switch val.(type) {
		case User:
			data = val.(User)
		default:
			fmt.Printf("Not User type. Ignore")
			continue
		}

		wg.Add(1)
		go func(data User, out chan<- interface{}, quotaCh chan struct{}) {
			defer wg.Done()
			quotaCh <- struct{}{}
			defer func() { <-quotaCh }()
			messages, err := GetMessages(data)
			if err != nil {
				fmt.Printf("ERROR")
			}
			out <- messages
		}(data, out, selectQuotaCh)
	}
	wg.Wait()
}

func checkSpamWorker(data []MsgID, out chan<- interface{}, wg *sync.WaitGroup, quotaCh chan struct{}) {
	defer wg.Done()
	quotaCh <- struct{}{}
	defer func() { <-quotaCh }()

	for _, msgID := range data {
		hasSpam, err := HasSpam(msgID)
		if err != nil {
			fmt.Printf("ERROR")
			continue
		}
		out <- MsgData{
			ID:      msgID,
			HasSpam: hasSpam,
		}
		runtime.Gosched()
	}

}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData

	wg := &sync.WaitGroup{}
	spamQuotaCh := make(chan struct{}, 5)

	for val := range in {
		var data []MsgID
		switch val.(type) {
		case []MsgID:
			data = val.([]MsgID)
		default:
			fmt.Printf("Not MsgID type. Ignore")
			continue
		}
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go checkSpamWorker(data, out, wg, spamQuotaCh)
		}
	}
	wg.Done()
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
	combinedResults := make([]MsgData, 0)

	for val := range in {
		switch val.(type) {
		case MsgData:
			data := val.(MsgData)
			combinedResults = append(combinedResults, data)
		default:
			fmt.Printf("Not MsgData type. Ignore")
			continue
		}
	}

	sort.Slice(combinedResults, func(i, j int) bool {
		left := strconv.FormatBool(combinedResults[i].HasSpam)
		right := strconv.FormatBool(combinedResults[j].HasSpam)
		if left > right {
			return true
		} else if left < right {
			return false
		}
		return combinedResults[i].ID < combinedResults[j].ID
	})

	for _, res := range combinedResults {
		out <- strconv.FormatBool(res.HasSpam) + " " + strconv.FormatUint(uint64(res.ID), 10)
	}
}
