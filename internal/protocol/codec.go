package protocol

func Encode(msg *Message) ([]byte, error) {
	return []byte{}, nil
}

func Decode(data []byte) (*Message, error) {
	return &Message{}, nil
}
