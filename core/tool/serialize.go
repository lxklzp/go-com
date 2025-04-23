package tool

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"go-com/config"
	"io"
)

const SerializeGzipHeaderLen = 20

// Serialize 序列化
func Serialize(encode func(encoder *gob.Encoder) error, isCompress bool) ([]byte, error) {
	bsBuf := config.BufPool.Get().(*bytes.Buffer)
	defer func() {
		bsBuf.Reset()
		config.BufPool.Put(bsBuf)
	}()

	// gob方式序列化
	encoder := gob.NewEncoder(bsBuf)
	err := encode(encoder)
	if err != nil {
		return nil, err
	}
	binaryData := bsBuf.Bytes()
	if isCompress {
		bsBuf.Reset()

		// 增加头部，处理gzip文件头问题
		for i := 0; i < SerializeGzipHeaderLen; i++ {
			binaryData = append(binaryData, 0)
		}
		copy(binaryData[SerializeGzipHeaderLen:], binaryData)
		for i := 0; i < SerializeGzipHeaderLen; i++ {
			binaryData[i] = byte(i)
		}

		// gzip压缩
		w := gzip.NewWriter(bsBuf)
		if _, err = w.Write(binaryData); err != nil {
			return nil, err
		}
		w.Flush()
		w.Close()
		binaryData = bsBuf.Bytes()
	}

	return binaryData, nil
}

// UnSerialize 反序列化
func UnSerialize(binaryData []byte, decode func(decoder *gob.Decoder) error, isCompress bool) error {
	bsBuf := config.BufPool.Get().(*bytes.Buffer)
	defer func() {
		bsBuf.Reset()
		config.BufPool.Put(bsBuf)
	}()

	if isCompress {
		// gzip解压缩
		bsBuf.Write(binaryData)
		r, err := gzip.NewReader(bsBuf)
		if err != nil {
			return err
		}
		binaryData, err = io.ReadAll(r)
		if err != nil {
			return err
		}

		// 删除头部，处理gzip文件头问题
		binaryData = binaryData[SerializeGzipHeaderLen:]
		bsBuf.Reset()
	}

	// gob方式反序列化
	bsBuf.Write(binaryData)
	decoder := gob.NewDecoder(bsBuf)
	return decode(decoder)
}
