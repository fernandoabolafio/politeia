package main

import (
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"path/filepath"

	"github.com/decred/politeia/politeiawww/backup"
)

func main() {
	var err error
	var reply backup.BackupDbReply

	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}

	doBackup := backup.BackupDbRequest{}

	err = client.Call("BackupServer.BackupDatabase", doBackup, &reply)
	if err != nil {
		log.Fatal("Problem backing up server: ", err)
	}

	root := "/Users/fernandoabolafio/Desktop/backup"
	// log.Println(reply)
	for _, file := range reply.Files {
		log.Printf("saving file %v", file.Name)
		filepath := filepath.Join(root, file.Name)
		_, err := os.Create(filepath)
		if err != nil {
			log.Fatal("couldn't create file", err)
		}

		err = ioutil.WriteFile(filepath, file.Payload, 0644)
		if err != nil {
			log.Fatal("couldn't save file", err)
		}
	}

}
