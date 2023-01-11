package structures

func StringEndsWith(mainString string, ending string) bool {
	if len(mainString) < len(ending) {
		return false
	}
	return mainString[len(mainString)-len(ending):] == ending
}
