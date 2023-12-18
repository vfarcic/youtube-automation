package main

func main() {
	// TODO: Publish binary
	// TODO: Publish as brew
	getArgs()
	choices := Choices{}
	for {
		choices.ChooseIndex()
	}
	// azureOpenAI := AzureOpenAI{}
	// responses, err := azureOpenAI.Chat("write a title for a youtube video about kubernetes dependencies and deployment ordering")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for _, resp := range responses {
	// 	fmt.Println(resp)
	// }
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
