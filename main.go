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

const (
	sqlCreateTable      = "CREATE TABLE mylist (item string, priority string);"
	sqlTables           = "SELECT name FROM sqlite_schema WHERE type ='table' AND name NOT LIKE 'sqlite_%';"
	sqlMyList           = "SELECT item, priority from mylist;"
	sqlMyListByPriority = "SELECT item, priority from mylist ORDER BY priority asc;"
	sqlCountMyItems     = "SELECT COUNT(*) from mylist;"
	sqlInsertMyList     = "INSERT INTO \"mylist\" (item, priority) VALUES ($f1, $f2);"
	sqlNowHere          = "SELECT datetime('now','-1 day','localtime');"
)

// Using a Pool to execute SQL in a concurrent HTTP handler.
func main() {
	fmt.Println("render-my-list ---- startup")
	//Get all env variables
	fmt.Println("LOCAL ENVIRONMENT")
	fmt.Println(os.Environ())
	fmt.Println("END LOCAL ENVIRONMENT")

	var err error
	dbpool, err = sqlitex.Open("file::memory:?cache=shared", 0, 10)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.TODO()
	conn := dbpool.Get(ctx)
	if conn == nil {
		fmt.Println("Unable to create conn from dbpool")
		return
	}
	// Execute a query.
	// Can I inject the "app name" into the table name - like "foo123-mylist"?
	fmt.Println("Initializing new in memory 'mylist' TABLE")

	if err := sqlitex.Execute(conn, sqlCreateTable, nil); err != nil {
		// handle err
		fmt.Println("Error trying to create table")
		fmt.Println(err)
	}

	executeSql(sqlTables, conn, nil)
	defer dbpool.Put(conn)

	http.HandleFunc("/", handle)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func executeSql(sql string, conn *sqlite.Conn, w http.ResponseWriter) {

	fmt.Printf("executeSql === sql:%s\n", sql)
	stmt, err := conn.Prepare(sql)
	if err != nil {
		fmt.Println(err)
		fmt.Fprint(w, err)
	}
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			fmt.Println(err)
			if w != nil {
				fmt.Fprint(w, err)
			}
		}
		if !hasRow {
			break
		}
		for i := 0; i < stmt.ColumnCount(); i++ {
			colName := stmt.ColumnName(i)
			fmt.Print(colName + ":" + stmt.GetText(colName))
			if w != nil {
				fmt.Fprint(w, colName+":"+stmt.GetText(colName))
			}
		}
		fmt.Println()
		if w != nil {
			fmt.Fprintln(w)
		}

	}

}

func handle(w http.ResponseWriter, r *http.Request) {
	item := r.FormValue("i")
	sort := r.FormValue("s")
	priority := r.FormValue("p")
	fmt.Printf("Adding new (sort=%s) priority:%s item:%s\n", sort, priority, item)
	conn := dbpool.Get(r.Context())
	if conn == nil {
		return
	}
	if item != "" {
		addItem(conn, item, priority, w)
	} else {
		response(conn, w, sort)
	}
	defer dbpool.Put(conn)
}
func addItem(conn *sqlite.Conn, item string, priority string, w http.ResponseWriter) {
	var err error
	stmt, err := conn.Prepare(sqlInsertMyList)
	if err != nil {
		fmt.Println(err)
		fmt.Fprint(w, err)
		return
	}
	stmt.SetText("$f1", item)
	stmt.SetText("$f2", priority)
	hasRow, err := stmt.Step()
	if err != nil {
		fmt.Println(err)
		fmt.Fprint(w, err)
	}
	if hasRow {
		fmt.Println("hasRow???? IDK ----->>>> ")
		fmt.Fprint(w, "hasRow???? IDK ----->>>> ")

	}

	if err := stmt.Finalize(); err != nil {
		fmt.Println(err)
		fmt.Fprint(w, err)
	}
}

func response(conn *sqlite.Conn, w http.ResponseWriter, sort string) {
	err := sqlitex.ExecuteTransient(conn, sqlNowHere, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			messageBack := fmt.Sprintf("Hello! Here's your list (%s)\n", stmt.ColumnText(0))
			fmt.Println(messageBack)
			fmt.Fprint(w, messageBack)
			sql := sqlMyList
			if sort!="" {
				sql = sqlMyListByPriority
			}
			executeSql(sql, conn, w)
			return nil
		},
	})

	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w, err)
	}
}
