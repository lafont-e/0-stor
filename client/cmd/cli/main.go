package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/zero-os/0-stor/client"
	"github.com/zero-os/0-stor/client/config"
)

func main() {
	if len(os.Args) != 4 && len(os.Args) != 5 {
		log.Println("usage:")
		log.Println("./cli conf_file upload file_name")
		log.Println("./cli conf_file download key result_file_name")
		return
	}
	confFile := os.Args[1]
	command := os.Args[2]

	// read config
	f, err := os.Open(confFile)
	if err != nil {
		log.Fatal(err)
	}

	conf, err := config.NewFromReader(f)
	if err != nil {
		log.Fatal(err)
	}

	// create client
	c, err := client.New(conf)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	switch command {
	case "upload":
		fileName := os.Args[3]
		err = uploadFile(c, fileName)
	case "download":
		key := os.Args[3]
		resultFile := os.Args[4]
		err = downloadFile(c, key, resultFile)
	}
	if err != nil {
		log.Fatalf("%v failed: %v", command, err)
	}
	log.Println("OK")
}

func uploadFile(c *client.Client, fileName string) error {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	log.Printf("0-stor key = %v\n", fileName)
	return c.Store([]byte(fileName), b)
}

func downloadFile(c *client.Client, key, resultFile string) error {
	b, err := c.Get([]byte(key))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(resultFile, b, 0666)
}