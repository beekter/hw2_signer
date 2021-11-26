package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type Sh struct {
	Data   []string
	Md5    []string
	Crc1   []string
	Crc2   []string
	Result []string
}

type ThreadResult struct {
	Data   string
	Th     []string
	Result string
}

type HashPrint struct {
	Sh
	Mh       []ThreadResult
	Combined string
}

var resultPrint HashPrint

func ExecutePipeline(jobs ...job) {
	var chans []chan interface{}
	jobCount := len(jobs)
	for i := 0; i < jobCount-1; i++ {
		chans = append(chans, make(chan interface{}))
	}

	pipelineWg := &sync.WaitGroup{}
	for i := 0; i < jobCount; i++ {
		if i == 0 {
			runFirstJob(pipelineWg, jobs[i], chans[i])
			continue
		}
		if i == jobCount-1 {
			runLastJob(pipelineWg, jobs[i], chans[i-1])
			continue
		}
		runJob(pipelineWg, jobs[i], chans[i-1], chans[i])
	}

	pipelineWg.Wait()
	PrettyPrint(resultPrint)
	//PrintResult(resultPrint)
}

func PrettyPrint(HashPrint HashPrint) {
	pp, err := json.MarshalIndent(HashPrint, "", "  ")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(pp), "asd")
}

func kPrintResult(HashPrint HashPrint) {
	for i := range HashPrint.Data {
		data := HashPrint.Data[i]
		fmt.Printf("%v SingleHash Data %v\n", data, data)
		fmt.Printf("%v SingleHash Md5(Data) %v\n", data, HashPrint.Md5[i])
		fmt.Printf("%v SingleHash crc32(Md5(Data)) %v\n", data, HashPrint.Crc1[i])
		fmt.Printf("%v SingleHash crc32(Data) %v\n", data, HashPrint.Crc2[i])
		fmt.Printf("%v SingleHash Result %v\n", data, HashPrint.Sh.Result[i])
		shResult := HashPrint.Sh.Result[i]
		for thi, thResult := range HashPrint.Mh[i].Th {
			fmt.Printf("%v MultiHash: crc32(Th+step1)) %v %v\n", shResult, thi, thResult)
		}
		fmt.Printf("%v MultiHash Result: %v\n\n", shResult, HashPrint.Mh[i].Result)
	}
	fmt.Printf("CombineResults %v", HashPrint.Combined)
}

func runFirstJob(pipelineWg *sync.WaitGroup, job job, out chan interface{}) {
	pipelineWg.Add(1)
	go func() {
		defer pipelineWg.Done()
		dummyIn := make(chan interface{})
		job(dummyIn, out)
		close(out)
	}()
}

func runLastJob(pipelineWg *sync.WaitGroup, job job, in chan interface{}) {
	pipelineWg.Add(1)
	go func() {
		defer pipelineWg.Done()
		dummyOut := make(chan interface{})
		job(in, dummyOut)
		close(dummyOut)
	}()
}

func runJob(pipelineWg *sync.WaitGroup, job job, in, out chan interface{}) {
	pipelineWg.Add(1)
	go func() {
		pipelineWg.Done()
		job(in, out)
	}()
}

func CombineResults(in chan interface{}, out chan interface{}) {
	var result string
	for data := range in {
		result += "_" + data.(string)
	}
	resultPrint.Combined = result[1:]
	out <- resultPrint.Combined
	close(out)
}

func MultiHash(in chan interface{}, out chan interface{}) {
	mhWg := sync.WaitGroup{}
	mhMu := &sync.Mutex{}
	for data := range in {
		fmt.Println(data)
		mhWg.Add(1)
		thResult := ThreadResult{
			Data:   data.(string),
			Th:     []string{"", "", "", "", "", ""},
			Result: "",
		}
		go func(data interface{}) {
			defer mhWg.Done()
			dataString, _ := data.(string)
			threads := []string{"0", "1", "2", "3", "4", "5"}

			thWg := sync.WaitGroup{}
			for i, th := range threads {
				thWg.Add(1)
				go func(i int, th string) {
					defer thWg.Done()
					thResult.Th[i] = DataSignerCrc32(th + dataString)
				}(i, th)
			}

			thWg.Wait()
			thResult.Result = strings.Join(thResult.Th, "")

			out <- thResult.Result
			mhMu.Lock()
			resultPrint.Mh = append(resultPrint.Mh, thResult)
			mhMu.Unlock()
		}(data)
	}
	mhWg.Wait()
	close(out)
}

func SingleHash(in chan interface{}, out chan interface{}) {
	singleHashWg := &sync.WaitGroup{}
	shMu := sync.Mutex{}

	inputDataChannel := make(chan interface{})
	inputDataChannel2 := make(chan interface{})
	md5Channel := make(chan interface{})
	crc1Channel := make(chan interface{})
	crc2Channel := make(chan interface{})

	singleHashWg.Add(1)
	go func() {
		defer singleHashWg.Done()
		for data := range in {
			dataString := strconv.Itoa(data.(int))
			shMu.Lock()
			resultPrint.Sh.Data = append(resultPrint.Sh.Data, dataString)
			shMu.Unlock()
			inputDataChannel <- dataString
			inputDataChannel2 <- dataString
		}
		close(inputDataChannel)
		close(inputDataChannel2)
	}()

	singleHashWg.Add(1)
	go func() {
		defer singleHashWg.Done()
		for data := range inputDataChannel {
			md5 := DataSignerMd5(data.(string))
			shMu.Lock()
			resultPrint.Sh.Md5 = append(resultPrint.Sh.Md5, md5)
			shMu.Unlock()
			md5Channel <- md5
		}
		close(md5Channel)
	}()

	handleCrc(singleHashWg, md5Channel, crc1Channel, 1, &resultPrint.Sh)
	handleCrc(singleHashWg, inputDataChannel2, crc2Channel, 2, &resultPrint.Sh)

	singleHashWg.Add(1)
	go func() {
		defer singleHashWg.Done()
		for crc1Data := range crc1Channel {
			crc2Data := <-crc2Channel
			result := crc2Data.(string) + "~" + crc1Data.(string)
			shMu.Lock()
			resultPrint.Sh.Result = append(resultPrint.Sh.Result, result)
			shMu.Unlock()
			out <- result
		}
	}()

	singleHashWg.Wait()
	close(out)
}

func handleCrc(wg *sync.WaitGroup, in, out chan interface{}, i int, resultPrint *Sh) {
	wg.Add(1)
	crcMutex := &sync.Mutex{}
	go func() {
		defer wg.Done()
		crcWg := sync.WaitGroup{}
		for inData := range in {
			crcWg.Add(1)
			go func(inData interface{}) {
				defer crcWg.Done()
				crcData := DataSignerCrc32(inData.(string))
				if i == 1 {
					crcMutex.Lock()
					resultPrint.Crc1 = append(resultPrint.Crc1, crcData)
					crcMutex.Unlock()
				} else {
					crcMutex.Lock()
					resultPrint.Crc2 = append(resultPrint.Crc2, crcData)
					crcMutex.Unlock()
				}
				out <- crcData
			}(inData)
		}
		crcWg.Wait()
		close(out)
	}()
}
