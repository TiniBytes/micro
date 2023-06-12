package protocol

type Response struct {
	HeadLength uint32
	BodyLength uint32
	MessageID  uint32
	Version    uint8
	Compress   uint8
	Serializer uint8
	Error      []byte
	Data       []byte
}
