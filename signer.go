package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	chIn := make(chan interface{})
	chOut := make(chan interface{})
	wg := &sync.WaitGroup{}
	for _, job := range jobs {
		wg.Add(1)

		jobGo := job
		chInGo := chIn
		chOutGo := chOut
		go func() {
			jobGo(chInGo, chOutGo)
			wg.Done()
			close(chOutGo)
		}()
		chIn = chOutGo
		chOut = make(chan interface{})
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	mu := sync.Mutex{}
	wgg := &sync.WaitGroup{}
	for data := range in {
		wgg.Add(1)
		dataGo := data
		go func() {
			dataInt, ok := dataGo.(int)
			if !ok {
				fmt.Println("Error type convert  in SingleHash")
				return
			}

			var data1, data2 string

			wg := &sync.WaitGroup{}
			wg.Add(2)

			go func() {
				data1 = DataSignerCrc32(strconv.Itoa(dataInt))
				wg.Done()
			}()

			go func() {
				mu.Lock()
				data2 = DataSignerMd5(strconv.Itoa(dataInt))
				mu.Unlock()
				data2 = DataSignerCrc32(data2)
				wg.Done()
			}()

			wg.Wait()

			out <- data1 + "~" + data2

			wgg.Done()
		}()
	}

	wgg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wgg := &sync.WaitGroup{}
	for data := range in {
		wgg.Add(1)
		dataGo := data
		go func() {
			dataStr, ok := dataGo.(string)
			if !ok {
				fmt.Println("Error type convert in MultiHash")
				return
			}

			wg := &sync.WaitGroup{}
			mhSlice := make([]string, 6)
			mu := sync.Mutex{}
			for i := 0; i < 6; i++ {
				wg.Add(1)
				j := i
				go func() {
					res := DataSignerCrc32(strconv.Itoa(j) + dataStr)

					mu.Lock()
					mhSlice[j] = res
					mu.Unlock()

					wg.Done()
				}()
			}
			wg.Wait()

			out <- strings.Join(mhSlice, "")
			wgg.Done()
		}()
	}

	wgg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var dataSlice []string

	for data := range in {
		dataStr, ok := data.(string)
		if !ok {
			fmt.Println("Error type convert in CombineResults")
			return
		}

		dataSlice = append(dataSlice, dataStr)
	}

	sort.Slice(dataSlice, func(i int, j int) bool {
		if dataSlice[i] < dataSlice[j] {
			return true
		}

		return false
	})

	out <- strings.Join(dataSlice, "_")
}
