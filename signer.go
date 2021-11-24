package main

import (
	"fmt"
	"strconv"
	"strings"
)

type chanMap map[int]chan interface{}

func ExecutePipeline(jobs ...job) {
	jobCount := len(jobs)

	if jobCount < 2 {
		return
	}

	chans := make(map[int]chan interface{})
	for i := 0; i < jobCount-1; i++ {
		chans[i] = make(chan interface{})
	}
	//chans[jobCount-1] = make(chan interface{}, 2)

	for i := 0; i < jobCount; i++ {
		if i == 0 {
			go runFirstJob(jobs[i], chans[i])
			continue
		}
		if i == jobCount-1 {
			go runLastJob(jobs[i], chans[i-1])
			continue
		}
		go jobs[i](chans[i-1], chans[i])
	}

	//fmt.Scanln();

	//for _, ch := range chans {
	//	close(ch)
	//}
}

func runFirstJob(firstJob job, out chan interface{}) {
	in := make(chan interface{})

	firstJob(in, out)

	close(in)
}

func runLastJob(lastJob job, in chan interface{}) {
	out := make(chan interface{})

	lastJob(in, out)

	close(out)
}

func CombineResults(in chan interface{}, out chan interface{}) {
	var result string
	for data := range in {
		result += "_" + data.(string)
	}
	result = result[1:]
	fmt.Println()
	fmt.Printf("CombineResults %v\n", result)
	out <- result
	close(out)
}

func MultiHash(in chan interface{}, out chan interface{}) {
	for data := range in {
		dataString, _ := data.(string)
		mhItems := []string{
			DataSignerCrc32("0" + dataString),
			DataSignerCrc32("1" + dataString),
			DataSignerCrc32("2" + dataString),
			DataSignerCrc32("3" + dataString),
			DataSignerCrc32("4" + dataString),
			DataSignerCrc32("5" + dataString),
		}
		result := strings.Join(mhItems[:], "")

		fmt.Println()
		for i, mhItem := range mhItems {
			fmt.Printf("%v MultiHash: crc32(th+step1)) %v %v\n", dataString, i, mhItem)
		}
		fmt.Printf("%v MultiHash: result: %v\n", dataString, result)

		out <- result
	}
	close(out)
}

func SingleHash(in chan interface{}, out chan interface{}) {
	for data := range in {
		dataString := strconv.Itoa(data.(int))
		md5String := DataSignerMd5(dataString)
		crc32Md5String := DataSignerCrc32(md5String)
		crc32String := DataSignerCrc32(dataString)
		result := crc32String + "~" + crc32Md5String

		fmt.Println()
		fmt.Printf("%v SingleHash data %v\n", data, dataString)
		fmt.Printf("%v SingleHash md5(data) %v\n", data, md5String)
		fmt.Printf("%v SingleHash crc32(md5(data)) %v\n", data, crc32Md5String)
		fmt.Printf("%v SingleHash crc32(data) %v\n", data, crc32String)
		fmt.Printf("%v SingleHash result %v\n", data, result)

		out <- result
	}
	close(out)
}
