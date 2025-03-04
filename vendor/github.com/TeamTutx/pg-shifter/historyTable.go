package shifter

import (
	"fmt"

	"github.com/TeamTutx/pg-shifter/util"
	"github.com/TeamTutx/plib/perror"
	"github.com/go-pg/pg/v10"
)

//Create history table
func (s *Shifter) createHistory(tx *pg.Tx, tableName string) (err error) {
	if s.isSkip(tableName) == false {
		historyTable := util.GetHistoryTableName(tableName)
		if tableExists := tableExists(tx, historyTable); tableExists == false {
			if err = s.execHistoryTable(tx, tableName, historyTable); err == nil {
				if err = s.dropHistoryConstraint(tx, historyTable); err == nil {
					err = s.createTrigger(tx, tableName)
				}
			}
		}
	}
	return
}

//dropHistory will drop history table
func (s *Shifter) dropHistory(tx *pg.Tx, tableName string, cascade bool) (err error) {
	historyTable := util.GetHistoryTableName(tableName)
	if tableExists := tableExists(tx, historyTable); tableExists == true {
		err = execTableDrop(tx, historyTable, cascade)
	}
	return
}

//dropHistoryConstraint will drop history table constraints
func (s *Shifter) dropHistoryConstraint(tx *pg.Tx, historyTable string) (err error) {
	sql := `
		ALTER TABLE %v DROP COLUMN IF EXISTS updated_at;
		ALTER TABLE %v ADD COLUMN IF NOT EXISTS created_at timetz DEFAULT now();`
	sql = fmt.Sprintf(sql, historyTable, historyTable)
	if _, err = tx.Exec(sql); err != nil {
		msg := `History Table Error: ` + historyTable + `
		SQL:` + sql
		err = perror.HTTPError(err, msg) // TODO : Replace with ExecError()
		fmt.Println(msg, err)
	}
	return
}

//execHistoryTable will execute history table creation
func (s *Shifter) execHistoryTable(tx *pg.Tx, tableName, historyTable string) (err error) {

	sql := `
	CREATE TABLE %v (
		id BIGSERIAL PRIMARY key,
		action VARCHAR(20),
		LIKE %v
	);
	`
	sql = fmt.Sprintf(sql, historyTable, tableName)
	if _, err = tx.Exec(sql); err != nil {
		msg := fmt.Sprintf("Table: %v", tableName)
		err = perror.HTTPError(err, msg) // TODO : Replace with ExecError()
		fmt.Println("History Error:", msg, err)
	}
	return
}

//IsSkip will check table contain skip tags
func (s *Shifter) isSkip(tableName string) (flag bool) {
	tableModel, isValid := s.table[tableName]
	if isValid {
		flag = util.SkipTag(tableModel)
	}
	return
}
