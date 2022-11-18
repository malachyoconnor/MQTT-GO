package packets

func CreateByteInline(input_binary []byte) byte {
	res, _ := CreateByte(input_binary)
	return res
}

// Take an array of 0s and 1s and return the byte representation
func CreateByte(input_binary []byte) (byte, bool) {
	if len(input_binary) > 8 {
		return 0, false
	}
	var result byte = 0
	for i, v := range input_binary {
		if !(v == 0 || v == 1) {
			return 0, false
		}
		result += (v << (7 - i))
	}
	return result, true
}
