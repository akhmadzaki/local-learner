package helpers

func ToInterfaceSlice[T any](src []T) []interface{} {
	dst := make([]interface{}, len(src))
	for i, v := range src {
		dst[i] = v
	}
	return dst
}
