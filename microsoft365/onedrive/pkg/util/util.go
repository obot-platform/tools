package util

func Deref[T any](v *T) (r T) {
	if v != nil {
		return *v
	}
	return
}
