package util

import (
	"errors"
	"fmt"
	"strings"
)

//GetTrigger : Get triggers by table name
func GetTrigger(tableName string) (trigger string) {
	var (
		aInsertTrigger string
		bUpdateTrigger string
		aUpdateTrigger string
		aDeleteTrigger string
	)
	// aInsertTrigger := getInsertTrigger(tableName)
	bUpdateTrigger, aUpdateTrigger = getUpdateTrigger(tableName)
	aDeleteTrigger = getDeleteTrigger(tableName)
	trigger = aInsertTrigger + bUpdateTrigger + aUpdateTrigger + aDeleteTrigger
	return
}

//Get after insert trigger
func getInsertTrigger(tableName string) (aInsertTrigger string) {
	if dbModel, valid := TableMap[tableName]; valid {
		if fields, values, _, _, err := getHistoryFields(dbModel, "NEW", "insert"); err == nil {
			aInsertTrigger = getAfterInsertTrigger(tableName, fields, values)
		} else {
			fmt.Println("getInsertTrigger: ", err.Error())
		}
	}
	return
}

//Get after insert trigger function and trigger by table name
func getAfterInsertTrigger(tableName, fields, values string) (aInsertTrigger string) {
	historyTable := tableName + HISTORY_TAG
	afterInsertTable := tableName + "_after_insert"
	fnQuery := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %v()
		RETURNS trigger AS
		$$
	    	BEGIN
	        	INSERT INTO %v (%v) 
	        	VALUES(%v);
	        	RETURN NEW;
	    	END;
		$$
		LANGUAGE 'plpgsql';
		`, afterInsertTable, historyTable, fields, values)
	triggerQuery := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %v ON %v;
		CREATE TRIGGER %v
		AFTER INSERT ON %v 
		FOR EACH ROW
		EXECUTE PROCEDURE %v();
		`, afterInsertTable, tableName, afterInsertTable, tableName, afterInsertTable)
	aInsertTrigger = fnQuery + triggerQuery
	return
}

//Get before and after update triggers
func getUpdateTrigger(tableName string) (bUpdateTrigger, aUpdateTrigger string) {
	if dbModel, valid := TableMap[tableName]; valid {
		if fields, values, updateCondition, updatedAt, err :=
			getHistoryFields(dbModel, "OLD", "update"); err == nil {
			if aUpdateTrigger = getAfterUpdateTrigger(tableName, fields,
				values, updateCondition); updatedAt {
				bUpdateTrigger = getBeforeUpdateTrigger(tableName)
			}
		} else {
			fmt.Println("getUpdateTrigger: ", err.Error())
		}
	}
	return
}

//Get after update trigger function and trigger by table name
func getAfterUpdateTrigger(tableName, fields, values,
	updateCondition string) (aUpdateTrigger string) {
	historyTable := tableName + HISTORY_TAG
	afterUpdateTable := tableName + "_after_update"
	fnQuery := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %v()
		RETURNS trigger AS
		$$
	    	BEGIN
	    		IF %v THEN
		        	INSERT INTO %v (%v) 
		        	VALUES(%v);
	        	END IF;
	        	RETURN NEW;
	    	END;
		$$
		LANGUAGE 'plpgsql';
		`, afterUpdateTable, updateCondition, historyTable, fields, values)
	triggerQuery := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %v ON %v;
		CREATE TRIGGER %v
		AFTER UPDATE ON %v 
		FOR EACH ROW
		EXECUTE PROCEDURE %v();
		`, afterUpdateTable, tableName, afterUpdateTable, tableName, afterUpdateTable)
	aUpdateTrigger = fnQuery + triggerQuery
	return
}

//Get before update trigger function and trigger by table name
func getBeforeUpdateTrigger(tableName string) (bUpdateTrigger string) {
	beforeUpdateTable := tableName + "_before_update"
	fnQuery := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %v()
		RETURNS trigger AS
		$$
	    	BEGIN
	        	NEW.updated_at = now();
	        	RETURN NEW;
	    	END;
		$$
		LANGUAGE 'plpgsql';
		`, beforeUpdateTable)
	triggerQuery := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %v ON %v;
		CREATE TRIGGER %v
		BEFORE UPDATE ON %v 
		FOR EACH ROW
		EXECUTE PROCEDURE %v();
		`, beforeUpdateTable, tableName, beforeUpdateTable, tableName, beforeUpdateTable)
	bUpdateTrigger = fnQuery + triggerQuery
	return
}

//Get after delete trigger
func getDeleteTrigger(tableName string) (aDeleteTrigger string) {
	if dbModel, valid := TableMap[tableName]; valid {
		if fields, values, _, _, err := getHistoryFields(dbModel, "OLD", "delete"); err == nil {
			aDeleteTrigger = getAfterDeleteTrigger(tableName, fields, values)
		} else {
			fmt.Println("getDeleteTrigger: ", err.Error())
		}
	}
	return
}

//Get after delete trigger function and trigger by table name
func getAfterDeleteTrigger(tableName, fields, values string) (aDeleteTrigger string) {
	historyTable := tableName + HISTORY_TAG
	afterDeleteTable := tableName + "_after_delete"
	fnQuery := fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %v()
		RETURNS trigger AS
		$$
	    	BEGIN
	        	INSERT INTO %v (%v) 
	        	VALUES(%v);
	        	RETURN OLD;
	    	END;
		$$
		LANGUAGE 'plpgsql';
		`, afterDeleteTable, historyTable, fields, values)
	triggerQuery := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS %v ON %v;
		CREATE TRIGGER %v
		AFTER DELETE ON %v 
		FOR EACH ROW
		EXECUTE PROCEDURE %v();
		`, afterDeleteTable, tableName, afterDeleteTable, tableName, afterDeleteTable)
	aDeleteTrigger = fnQuery + triggerQuery
	return
}

//Get history table fields from struct model of database
func getHistoryFields(dbModel interface{}, dataTag, action string) (
	fields string, values string, updateCondition string, updatedAt bool, err error) {
	fieldMap := GetStructField(dbModel)
	for _, inputField := range fieldMap {
		if tagValue, exists := inputField.Tag.Lookup("pg"); exists {
			curField := strings.Split(tagValue, ",")
			updatedAtExists := strings.Contains(curField[0], "updated_at")
			if len(curField) > 0 && !updatedAtExists {
				fields += curField[0] + ","
				if curField[0] == "created_at" {
					values += "NOW(),"
				} else {
					values += dataTag + "." + curField[0] + ","
					updateCondition += " OLD." + curField[0] + " <> NEW." + curField[0] + " OR"
				}
			} else if updatedAtExists {
				updatedAt = true
			}
		} else if !exists {
			err = errors.New("sql tag is missing in database struct model " + inputField.Name)
			return
		}
	}
	fields += "action"
	values += "'" + action + "'"
	updateCondition = strings.TrimSuffix(updateCondition, "OR")
	return
}
