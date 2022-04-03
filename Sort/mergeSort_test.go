package Sort

import (
	"fmt"
	"testing"
)

func TestMergeSort(t *testing.T) {
	arr := MergeSort([]int{5, 3, 2, 4, 1})
	fmt.Println(arr)
}
