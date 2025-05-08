package util

func Ptr[T any](v T) *T {
	return &v
}

func Deref[T any](v *T) (r T) {
	if v != nil {
		return *v
	}
	return
}
