package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	// 虚拟节点倍数为3， 方便测试使用自定义的哈希函数，输入一个string类型数字，直接转回int类型
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// 新增三个节点，由于虚拟节点倍数为3， 实际节点为 02、12、22； 04、14、24； 06；16；26
	hash.Add("6", "4", "2")

	// 根据上面的虚拟节点， 我们自己可以算出左边k对应的节点是哪个
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	// 对上面所有的key进行遍历，如果于我们求出结果不一致，说明一致性哈希设计出错
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	// 在原有的基础再新增一个节点8， 虚拟节点08、18、28
	hash.Add("8")

	// 那么前面测试用例中27原本对应02，现在最接近的是虚拟节点28，应该被映射到节点8
	testCases["27"] = "8"

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

}
