package ptr

func String(val string) *string {
	if val == "" {
		return nil
	}
	return &val
}
