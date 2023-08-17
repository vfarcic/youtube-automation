package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func readYaml(path string) Video {
	var video Video
	data, err := os.ReadFile(path)
	if err != nil {
		return video
	}
	err = yaml.Unmarshal(data, &video)
	if err != nil {
		log.Fatal(err)
	}
	return video
}

func writeYaml(video Video, path string) {
	data, err := yaml.Marshal(&video)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
