package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

var dbpool *sqlitex.Pool

// Using a Pool to execute SQL in a concurrent HTTP handler.
func main() {
	var err error
	dbpool, err = sqlitex.Open("file:memory:?mode=memory", 0, 10)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", handle)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	conn := dbpool.Get(r.Context())
	if conn == nil {
		return
	}
	defer dbpool.Put(conn)
	// Execute a query.
	var err error
	sqlNowHere := "SELECT datetime('now','-1 day','localtime');"
	err = sqlitex.ExecuteTransient(conn, sqlNowHere, &sqlitex.ExecOptions{
	  ResultFunc: func(stmt *sqlite.Stmt) error {
	    w.Header().Set("Content-Type","application/text")
	    messageBack := fmt.Sprintf("Hello! Its %s here how, what time is it for you?", stmt.ColumnText(0))
	    //fmt.Println(stmt.ColumnText(0))
	    //fmt.Fprintf(w, "Hello, %q", stmt.ColumnText(0))
	    fmt.Println(messageBack)
	    fmt.Fprint(w, messageBack)
	    return nil
	  },
	})
	if err != nil {
          fmt.Println(err)
	  //return err
	}
	/*
	stmt := conn.Prep("SELECT foo FROM footable WHERE id = $id;")
	stmt.SetText("$id", "_user_id_")
	for {
		if hasRow, err := stmt.Step(); err != nil {
			// ... handle error
		} else if !hasRow {
			break
		}
		foo := stmt.GetText("foo")
		// ... use foo
		fmt.Fprintln(w, foo)
	}
	*/
}

