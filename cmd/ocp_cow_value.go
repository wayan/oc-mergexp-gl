package cmd

func ocpCowValue[T any](sys System, ocpVal, cowVal T) T {
	if sys == Cow {
		return cowVal
	}
	return ocpVal
}
