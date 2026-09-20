package main

import "fmt"

func main(){
    fmt.Println(twoSum([]int{1, 2, 3, 4}, 5))
}

func twoSum(nums []int, target int) []int {
	for i := 0; i < len(nums) -1; i++{
		for j:= i + 1; j < (len(nums)); j++{
			//fmt.Println(i, j)
			if (nums[i] + nums[j]) == target{
				return []int{i, j}
			}
		}
	}
	return []int{}
}
