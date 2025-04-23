package hash

import (
	"github.com/spaolacci/murmur3"
)

/********** 常用的哈希算法

MurmurHash 是一种非加密型哈希算法，具有良好的性能、低碰撞率、优秀的雪崩效应（微小输入变化导致输出巨大变化）

xxHash 与MurmurHash特性相当，适用于处理大量数据的场景（文件数据校验等）。https://github.com/cespare/xxhash

**********/

func Hash(data []byte) uint64 {
	return murmur3.Sum64(data) % 10 // MurmurHash是一种非加密型哈希算法，具有良好的性能、低碰撞率、优秀的雪崩效应（微小输入变化导致输出巨大变化）
}
