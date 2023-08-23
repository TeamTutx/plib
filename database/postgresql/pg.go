package postgresql

import (
	"crypto/tls"
	"errors"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"gitlab.com/g-harshit/plib/conf"
	"gitlab.com/g-harshit/plib/constant"
	"gitlab.com/g-harshit/plib/perror"
)

var (
	//StartLogging is set to log query
	StartLogging     bool
	isDebuggerActive bool
	debuggerStatus   map[*pg.DB]bool
	dbPostgresWrite  map[string]*pg.DB
	dbPostgresRead   map[string][]*pg.DB
	masterContainer  = "database.master"
	slaveContainer   = "database.slaves"
)

func init() {
	debuggerStatus = make(map[*pg.DB]bool)
	dbPostgresWrite = make(map[string]*pg.DB)
	dbPostgresRead = make(map[string][]*pg.DB)
}

//init master connection
func initMaster() (err error) {
	if conf.Exists(masterContainer) {

		if dbPostgresWrite[masterContainer] == nil {
			var postgresWriteOption pg.Options
			if postgresWriteOption, err = getPostgresOptions(masterContainer); err == nil {
				dbPostgresWrite[masterContainer] = pg.Connect(&postgresWriteOption)
			}
		}
	} else {
		err = errors.New("master config does not exists")
	}
	return
}

//init slave connections
func initSlaves() (err error) {
	if conf.Exists(slaveContainer) {
		slaves := conf.StringSlice(slaveContainer, []string{})

		if dbPostgresRead[slaveContainer] == nil {
			dbPostgresRead[slaveContainer] = make([]*pg.DB, len(slaves))
		}
		for i, container := range slaves {
			if dbPostgresRead[slaveContainer][i] == nil {
				var postgresReadOption pg.Options
				if postgresReadOption, err = getPostgresOptions(container); err != nil {
					break
				}
				dbPostgresRead[slaveContainer][i] = pg.Connect(&postgresReadOption)
			}
		}
	} else {
		err = errors.New("slaves config does not exists")
	}
	return
}

//CreateMaster will create new master connection
func CreateMaster() (err error) {
	if conf.Exists(masterContainer) {
		var postgresWriteOption pg.Options
		if postgresWriteOption, err = getPostgresOptions(masterContainer); err == nil {
			dbPostgresWrite[masterContainer] = pg.Connect(&postgresWriteOption)
		}
	} else {
		err = errors.New("master config does not exists")
	}
	return
}

//CreateSlave will create new slave connections
func CreateSlave() (err error) {
	if conf.Exists(slaveContainer) {
		slaves := conf.StringSlice(slaveContainer, []string{})

		if dbPostgresRead[slaveContainer] == nil {
			dbPostgresRead[slaveContainer] = make([]*pg.DB, len(slaves))
		}
		for i, container := range slaves {

			var postgresReadOption pg.Options
			if postgresReadOption, err = getPostgresOptions(container); err != nil {
				break
			}
			dbPostgresRead[slaveContainer][i] = pg.Connect(&postgresReadOption)
		}
	} else {
		err = errors.New("slaves config does not exists")
	}
	return
}

//set postgres connection options from conf
func getPostgresOptions(container string) (pgOption pg.Options, err error) {
	if !conf.Exists(container) {
		err = errors.New("container for postgres configuration not found")
		return
	}
	host := conf.String(container+".host", "")
	port := conf.String(container+".port", "")
	addr := ""
	if host != "" && port != "" {
		addr = host + ":" + port
	}
	pgOption.Addr = addr
	pgOption.User = conf.String(container+".username", "")
	pgOption.Password = conf.String(container+".password", "")
	pgOption.Database = conf.String(container+".db", "")
	pgOption.MaxRetries = conf.Int(container+".maxRetries", 3)
	pgOption.RetryStatementTimeout = conf.Bool(container+".retryStmTimeout", false)
	if os.Getenv(constant.EnvironmentVariableEnv) == constant.StagingEnv {
		pgOption.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return
}

//Conn will return postgres connection
func Conn(writable bool) (dbConn *pg.DB, err error) {
	rand.Seed(time.Now().UnixNano())
	if writable {
		if err = initMaster(); err == nil {
			dbConn = dbPostgresWrite[masterContainer]
		}
	} else {
		if err = initSlaves(); err == nil {
			if len(dbPostgresRead) == 0 {
				if err = initMaster(); err == nil {
					dbConn = dbPostgresWrite[masterContainer]
				}
			} else {
				dbConn = dbPostgresRead[slaveContainer][rand.Intn(len(dbPostgresRead[slaveContainer]))]
			}
		}
	}
	if err == nil {
		if !debuggerStatus[dbConn] {
			debuggerStatus[dbConn] = true
			logQuery(dbConn)
		}
	} else {
		err = perror.ConnError(err)
	}
	return
}

//Tx will return postgres transaction
func Tx() (tx *pg.Tx, err error) {
	var conn *pg.DB
	if conn, err = Conn(true); err == nil {
		if tx, err = conn.Begin(); err != nil {
			err = perror.TxError(err)
		}
	}
	return
}

//ConnByContainer will return postgres connection by container
func ConnByContainer(container string) (*pg.DB, error) {
	if strings.HasSuffix(container, "master") {
		oldContainer := masterContainer
		masterContainer = container
		conn, err := Conn(true)
		masterContainer = oldContainer
		return conn, err
	} else if strings.HasSuffix(container, "slaves") {
		oldContainer := slaveContainer
		slaveContainer = container
		conn, err := Conn(false)
		slaveContainer = oldContainer
		return conn, err
	}
	return nil, errors.New("No master or slaves container found in: " + container)
}

//logQuery : Print postgresql query on terminal
func logQuery(conn *pg.DB) {
	if StartLogging {
		conn.AddQueryHook(Hook{})
	}
}

//Debug : Print postgresql query on terminal
func Debug(conn *pg.DB) {
	if !isDebuggerActive {
		isDebuggerActive = true
		conn.AddQueryHook(Hook{})
	}
}
