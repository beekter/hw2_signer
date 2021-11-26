package main

import (
	"fmt"
)

func main() {
	inputData := []int{0, 1}

	inputJob := job(func(in, out chan interface{}) {
		for _, fibNum := range inputData {
			out <- fibNum
		}
	})

	outputJob := job(func(in, out chan interface{}) {
		dataRaw := <-in
		data, ok := dataRaw.(string)
		if !ok {
			fmt.Println("cant convert Result Data to string")
			return
		}
		fmt.Println(data)
	})

	ExecutePipeline(
		inputJob,
		SingleHash,
		MultiHash,
		CombineResults,
		outputJob,
	)
}
