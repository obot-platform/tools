package sqlite_vec

import (
	"bytes"
	"encoding/binary"
)

func DeserializeFloat32(data []byte) ([]float32, error) {
	buf := bytes.NewReader(data)
	var vector []float32
	err := binary.Read(buf, binary.LittleEndian, &vector)
	if err != nil {
		return nil, err
	}
	return vector, nil
}
