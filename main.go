package main

func main() {
	getArgs()
	choices := Choices{}
	for {
		choices.ChooseIndex()
	}
}

// func deleteEmpty(s []string) []string {
// 	var r []string
// 	for _, str := range s {
// 		if str != "" {
// 			r = append(r, strings.TrimSpace(str))
// 		}
// 	}
// 	return r
// }
