package main

func main() {
	// TODO: Publish binary
	// TODO: Publish as brew
	getArgs()
	choices := Choices{}
	for {
		choices.ChooseIndex()
	}
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
