package packetcodec

import "encoding/json"

func Encode(packet any) ([]byte, error) {
	return json.Marshal(packet)
}

func Decode(data []byte, packet any) error {
	return json.Unmarshal(data, packet)
}
