package backup

import (
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/decred/politeia/politeiawww/database"
)

type File struct {
	Name    string
	Payload []byte
}

// BackupService provides the methods backing up the server
type BackupService interface {
	BackupDatabase(BackupDbRequest, *BackupDbRequest) error
}

// BackupServer is the server used only for backup
type BackupServer struct {
	db database.Database
}

// BackupDbRequest Command used to fetch the backup of the database
type BackupDbRequest struct{}

// BackupDbReply Command used to reply to the backup of the database
type BackupDbReply struct {
	Files []File
}

func convertFileFromDatabase(file database.File) File {
	return File{
		Name:    file.Name,
		Payload: file.Payload,
	}
}

// BackupDatabase is a method to execute the backup of the dabase and assign it to the provided
// backup reply
func (bs *BackupServer) BackupDatabase(breq BackupDbRequest, breply *BackupDbReply) error {
	files, err := bs.db.BackupUsersDatabase()
	if err != nil {
		return err
	}
	// log.Printf("got files %v", files)
	var reply BackupDbReply
	for _, f := range files {
		reply.Files = append(reply.Files, convertFileFromDatabase(f))
	}
	*breply = reply
	return nil
}

// InitBackupServer inits a rpc server for executing backup tasks
func InitBackupServer(db database.Database) {
	bs := new(BackupServer)
	bs.db = db

	err := rpc.Register(bs)
	if err != nil {
		log.Fatalf("Format of service BackupServer isn't correct", err)
	}

	rpc.HandleHTTP()
	// Listen to TPC connections on port 1234
	listener, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("Listen error: ", e)
	}

	log.Printf("Serving RPC server on port %d", 1234)
	// Start accept incoming HTTP connections
	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal("Error serving: ", err)
	}

}
