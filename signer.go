package main

import "fmt"

type chanMap map[int]chan interface{}

func ExecutePipeline(jobs ...job) {
	firstJob := jobs[0]
	lastJob := jobs[len(jobs)-1]

	in := make(chan interface{}, 20)
	out := make(chan interface{}, 20)

	go firstJob(in, out)

	fmt.Scanln()

	go lastJob(out, in)

	fmt.Scanln()
}

func CombineResults(in chan interface{}, out chan interface{}) {
	value := <-in
	fmt.Printf("Combining res: %v\n", value)
	out <- value
}

func MultiHash(in chan interface{}, out chan interface{}) {
	value := <-in
	fmt.Printf("MultiHash: %v\n", value)
	out <- value
	//data := <- in
	//dataString, _ := data.(string)
	//out <- DataSignerCrc32(dataString)+"~"+DataSignerCrc32(DataSignerMd5(dataString))
}

func SingleHash(in chan interface{}, out chan interface{}) {
	value := <-in
	fmt.Printf("SingleHash: %v\n", value)
	out <- value
	//data := <- in
	//dataString, _ := data.(string)
	//out <- DataSignerCrc32(dataString)+"~"+DataSignerCrc32(DataSignerMd5(dataString))
}
