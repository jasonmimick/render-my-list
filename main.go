package main

import (
	"context"
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

	conn := dbpool.Get(context.TODO())
	if conn == nil {
		return
	}
	defer dbpool.Put(conn)
	// Execute a query.
	// Can I inject the "app name" into the table name - like "foo123-mylist"?
	sql := "CREATE TABLE mylist (item string, priority string);"
	   fmt.Println("Initializing new in memory 'mylist' TABLE")
	err = sqlitex.ExecuteTransient(conn, sql, &sqlitex.ExecOptions{
	  ResultFunc: func(stmt *sqlite.Stmt) error {
	    fmt.Println("RESULT")
	    fmt.Println(stmt.ColumnText(0))
	    return nil
	  },
	})
	if err != nil {
	  fmt.Println("ERROR")
          fmt.Println(err)
	  //return err
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getMyList(conn *sqlite.Conn, w http.ResponseWriter) {


	stmt, err := conn.Prepare("SELECT item, priority from mylist;")
	if err != nil {
        	fmt.Println(err)
	  	fmt.Fprint(w,err)
	}
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			fmt.Println(err)
			fmt.Fprint(w,err)
		}
		if !hasRow {
			break
		}
		itemInfo := fmt.Sprintf("[%s]:%s", stmt.GetText("item"), stmt.GetText("priority"))
	    	fmt.Println( itemInfo )
	    	fmt.Fprint(w, itemInfo )
		
		
	}

}

func handle(w http.ResponseWriter, r *http.Request) {
 	//io.WriteString(res, "name: "+req.FormValue("name"))
    	//io.WriteString(res, "\nphone: "+req.FormValue("phone"))
 	item := r.FormValue("i")
	priority := r.FormValue("p")	
	fmt.Printf("Adding new item(%s): %s", priority, item)
	conn := dbpool.Get(r.Context())
	if conn == nil {
		return
	}
	defer dbpool.Put(conn)
	// Execute a query.
	var err error

	stmt, err := conn.Prepare("INSERT INTO mylist (item, priority) VALUES ($f1, $f2);")
	if err != nil {
		fmt.Println(err)
		fmt.Fprint(w,err)
		return
	}
	stmt.SetText("$f1", item)
	stmt.SetText("$f2", priority)
	//stmt.SetInt64("$f2", int64(i))
	hasRow, err := stmt.Step()
	if err != nil {
		fmt.Println(err)
		fmt.Fprint(w,err)
	}
	if hasRow {
		fmt.Println("hasRow???? IDK ----->>>> ")
		fmt.Fprint(w,"hasRow???? IDK ----->>>> ")

	}

	if err := stmt.Finalize(); err != nil {
		fmt.Println(err)
		fmt.Fprint(w,err)
	}

	sqlNowHere := "SELECT datetime('now','-1 day','localtime');"
	err = sqlitex.ExecuteTransient(conn, sqlNowHere, &sqlitex.ExecOptions{
	  ResultFunc: func(stmt *sqlite.Stmt) error {
	    w.Header().Set("Content-Type","application/text")
	    messageBack := fmt.Sprintf("Hello! Its %s here how, what time is it for you?", stmt.ColumnText(0))
	    //fmt.Println(stmt.ColumnText(0))
	    //fmt.Fprintf(w, "Hello, %q", stmt.ColumnText(0))
	    fmt.Println(messageBack)
	    fmt.Fprint(w, messageBack)
	    getMyList(conn, w)	    
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

