package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
	"gitlab.com/g-harshit/plib/ally"
)

func (d Hook) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	q.StartTime = time.Now()
	return c, nil
}

func (d Hook) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	fmt.Println("----DEBUGGER----")
	query, _ := q.FormattedQuery()
	fmt.Println(string(query))
	fmt.Printf("\n%v", ally.GetTraceFileWithoutDepth())
	fmt.Printf("\n%v\nErr - %v\n\n", time.Since(q.StartTime), q.Err)
	// TODO:
	// fmt.Printf("\nFile: %v : %v\nFunction: %v\nQuery Execution Taken: %s\n%s%s\n\n",
	// 	q.File, q.Line, q.Func, time.Since(q.StartTime), q.Err.Error())
	return nil
}
