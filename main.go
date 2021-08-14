package main

import (
	"fmt"
)

func main() {
	//inputData := []int{0, 1}
	inputData := []string{"one", "two"}
	var res string

	//var in, out chan interface{}

	inputJob := job(func(in, out chan interface{}) {
		fmt.Printf("in chan: %T\n", in)
		fmt.Printf("out chan: %T\n", out)
		for _, datum := range inputData {
			fmt.Printf("writing to out: %v\n", datum)
			out <- datum
		}
	})

	outputJob := job(func(in, out chan interface{}) {
		for dataRaw := range in {
			data := fmt.Sprintf("%v", dataRaw)
			res += " " + data
		}
	})

	ExecutePipeline(
		inputJob,
		outputJob,
	//job(MultiHash),
	//job(CombineResults),
	)

	fmt.Printf("Result: %v\n", res)
}
