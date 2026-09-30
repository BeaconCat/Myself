package auth

// 测试不需要生产强度的口令哈希：调低代价，免得整套用例慢几倍
func init() { ScryptN = 1 << 12 }
