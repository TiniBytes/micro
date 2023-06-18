package compress

type Compressor interface {
	Code() uint8
	Compress(data []byte) ([]byte, error)
	UnCompress(data []byte) ([]byte, error)
}
