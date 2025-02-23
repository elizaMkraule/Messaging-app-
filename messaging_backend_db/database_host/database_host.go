// This package defines database hosts and http methods that require their use. A database host is used to store databases.
package database_host

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/docAndColl"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/skiplist"
)

// Defines a database host struct which contains a skiplist of databases
type Database_host struct {
	Name        string
	Mu          sync.Mutex
	DatabaseMap map[string]*database.Database // Map of database names to database instances
	DBSkipList  skiplist.List[string, *database.Database]
}

// Constructs a new database_host
func NewDatabaseHost(name string) *Database_host {
	slog.Info("making new databasehost")
	return &Database_host{
		Name:       name,
		DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz"),
	}
}

// Takes in the name of a database and attempts to retrieve that database from the database host
// Returns the database and true on success, or an empty value and false on failure
func (db_host *Database_host) GetDatabase(databaseName string) (*database.Database, bool) {
	db_host.Mu.Lock()
	defer db_host.Mu.Unlock()

	slog.Info("start get db")
	db, exist := db_host.DBSkipList.Find(databaseName)
	if !exist {
		slog.Error("db doesnt exist")
		return nil, false
	}
	slog.Info("middle of get database")

	slog.Info("get database, from db_host, returning db and exist")

	return db, exist

}

// Takes in a writer and a database name and attempts to delete that database from the database host
// Writes the appropriate message to the header based on success/failure
func (db_host *Database_host) DeleteDatabase(w http.ResponseWriter, dbName string) {
	slog.Info("In delete database")
	slog.Info("DeleteDatabase: " + dbName)

	// SKIPLISTS:
	db, removed := db_host.DBSkipList.Remove(dbName)

	if !removed {
		slog.Info("did not remove database successfully")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`"not found"`))
	} else {
		slog.Info("removed database successfully, dbname: " + db.Name)
		w.WriteHeader(http.StatusNoContent)
	}

}

// Takes in information on a database and attempts to put the database into the database host
// Writes the appropriate header based on success/failure
func (db_host *Database_host) PutDatabaseIntoServer(owlDB *Database_host, w http.ResponseWriter, r *http.Request, name string) {
	slog.Info("server success")
	slog.Info(owlDB.Name)

	var newDatabase database.Database

	slog.Info("before .Name")

	//adding values to database
	newDatabase.Name = name

	slog.Info("after name")

	uri := map[string]string{
		"uri": r.URL.Path,
	}
	slog.Info("after uri")

	jsonData, err := json.MarshalIndent(uri, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to create database " + newDatabase.Name + ": url cannot be marshalled"))
	}
	slog.Info("after JSON")

	newDatabase.URI = jsonData
	slog.Info("starting skiplist, PutDatabaseIntoServer")

	// SKIPLISTS:

	newDatabase.DocSkipList = skiplist.NewList[string, *docAndColl.Document]("", "zzz")

	// first do the check for updating
	c := func(name string, db *database.Database, exists bool) (newValue *database.Database, err error) {
		// if the node already exists, we don't want to update it. We cannot replace an already existing DB. Just return an error
		// else, if the node does not exists, insert the database
		if exists {
			return nil, fmt.Errorf("error in check of PutDatabaseIntoServer, database already exists")
		} else {
			return &newDatabase, nil
		}
	}

	slog.Info("running Upsert")
	slog.Info(newDatabase.Name)

	updating, err := owlDB.DBSkipList.Upsert(newDatabase.Name, c)
	if err != nil {
		slog.Error("error after upsert in PutDocIntoCollection:", err)
	}

	if updating {
		slog.Info("PUT unable to create database " + newDatabase.Name + ": exists")
		w.WriteHeader(http.StatusBadRequest)
		msg := "unable to create database " + newDatabase.Name + ": exists"
		jsonMsg, _ := json.Marshal(msg)
		w.Write(jsonMsg)

	} else {
		slog.Info("PUT database: creating Database", "Name", newDatabase.Name)
		w.WriteHeader(http.StatusCreated)
		w.Write(newDatabase.URI)
	}
}
