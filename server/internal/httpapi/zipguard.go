package httpapi

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// 压缩包目录的体量上限：zip.OpenReader 会把整个中央目录读进内存（每条约 250 字节），
// 精心构造的包可以用几百 MB 塞进上千万条，公开的预览接口因此能被刷爆内存。打开前先读尾部目录记录核对。
const (
	zipMaxEntries = 200_000
	zipMaxDirSize = 16 << 20
)

var errZipTooBig = errors.New("zip directory too large")

// checkZipDirectory 读取 zip 尾部的中央目录结束记录（含 ZIP64），条目数或目录体积超限时返回 errZipTooBig。
// 不是合法 zip 时返回 nil，交给 zip.OpenReader 报错。
func checkZipDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return err
	}
	// EOCD 22 字节 + 最长 65535 字节注释
	tail := min(st.Size(), 22+65535)
	buf := make([]byte, tail)
	if _, err := f.ReadAt(buf, st.Size()-tail); err != nil && err != io.EOF {
		return err
	}
	i := len(buf) - 22
	for ; i >= 0; i-- {
		if binary.LittleEndian.Uint32(buf[i:]) == 0x06054b50 {
			break
		}
	}
	if i < 0 {
		return nil
	}
	entries := uint64(binary.LittleEndian.Uint16(buf[i+10:]))
	dirSize := uint64(binary.LittleEndian.Uint32(buf[i+12:]))
	// ZIP64：取紧挨在前面的 ZIP64 定位记录指向的 ZIP64 EOCD
	if entries == 0xffff || dirSize == 0xffffffff {
		if i >= 20 && binary.LittleEndian.Uint32(buf[i-20:]) == 0x07064b50 {
			off := int64(binary.LittleEndian.Uint64(buf[i-20+8:]))
			rec := make([]byte, 56)
			if off >= 0 && off+56 <= st.Size() {
				if _, err := f.ReadAt(rec, off); err == nil && binary.LittleEndian.Uint32(rec) == 0x06064b50 {
					entries = binary.LittleEndian.Uint64(rec[32:])
					dirSize = binary.LittleEndian.Uint64(rec[40:])
				}
			}
		}
	}
	if entries > zipMaxEntries || dirSize > zipMaxDirSize {
		return errZipTooBig
	}
	return nil
}
