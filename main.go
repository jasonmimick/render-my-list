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

var showTablesSql = "SELECT name FROM sqlite_schema WHERE type ='table' AND name NOT LIKE 'sqlite_%';"

var myListSql = "SELECT item, priority from mylist;" 
// Using a Pool to execute SQL in a concurrent HTTP handler.
func main() {
	var err error
	dbpool, err = sqlitex.Open("file::memory:?cache=shared", 0, 10)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", handle)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

    ctx := context.TODO()
	conn := dbpool.Get(ctx)
	if conn == nil {
        fmt.Println("Unable to create conn from dbpool")
		return
	}
	// Execute a query.
	// Can I inject the "app name" into the table name - like "foo123-mylist"?
	sql := "CREATE TABLE mylist (item string, priority string);"
	fmt.Println("Initializing new in memory 'mylist' TABLE")

    if err := sqlitex.Execute(conn, sql, nil); err != nil {
		// handle err
        fmt.Println("Error trying to create table")
        fmt.Println(err)
	}
    runMySql(showTablesSql,conn,nil)
	defer dbpool.Put(conn)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func runMySql(sql string, conn *sqlite.Conn, w http.ResponseWriter) {

    fmt.Printf("runMySql === sql:%s\n",sql)
	stmt, err := conn.Prepare(sql)
	if err != nil {
        	fmt.Println(err)
	  	fmt.Fprint(w,err)
	}
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			fmt.Println(err)
            if w!= nil {
			    fmt.Fprint(w,err)
		    }
        }
		if !hasRow {
			break
		}
        for i:=0; i < stmt.ColumnCount(); i++ {
            colName := stmt.ColumnName(i)
            fmt.Print( colName+":"+stmt.GetText(colName) )
            if w!= nil {
                fmt.Fprint(w, colName+":"+stmt.GetText(colName) )
            }
        }
        fmt.Println()
        if w!= nil { 
            fmt.Fprintln(w) 
        }
                
        	
	}

}

func handle(w http.ResponseWriter, r *http.Request) {
 	item := r.FormValue("i")
	priority := r.FormValue("p")	
	fmt.Printf("Adding new item(%s): %s\n", priority, item)
	conn := dbpool.Get(r.Context())
	if conn == nil {
		return
	}
    if item != "" {
        addItem(conn, item, priority, w)
    } else {
        fmt.Println("item was empty, not adding")
    }
    response(conn, w)
	defer dbpool.Put(conn)
}
func addItem(conn *sqlite.Conn,item string, priority string,w http.ResponseWriter) {
    var err error
	stmt, err := conn.Prepare("INSERT INTO \"mylist\" (item, priority) VALUES ($f1, $f2);")
	if err != nil {
		fmt.Println(err)
		fmt.Fprint(w,err)
		return
	}
	stmt.SetText("$f1", item)
	stmt.SetText("$f2", priority)
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
}

func response(conn *sqlite.Conn,w http.ResponseWriter) {
	sqlNowHere := "SELECT datetime('now','-1 day','localtime');"
    err := sqlitex.ExecuteTransient(conn, sqlNowHere, &sqlitex.ExecOptions{
	  ResultFunc: func(stmt *sqlite.Stmt) error {
	    //w.Header().Set("Content-Type","application/text")
	    messageBack := fmt.Sprintf("Hello! Here's your list (%s)\n", stmt.ColumnText(0))
	    fmt.Println(messageBack)
	    fmt.Fprint(w, messageBack)
        runMySql(myListSql, conn, w)
        
	    return nil
	  },
	})
	if err != nil {
        fmt.Println(err)
        fmt.Fprintln(w,err)
	}
}

