package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// LazyPdo is a custom db client
type LazyPdo struct {
	dsn      string
	user     string
	password string
	options  map[string]string
	commands []string
	pdo      *sql.DB
}

func NewLazyPdo(dsn string, user string, password string, options map[string]string) *LazyPdo {
	l := &LazyPdo{dsn, user, password, options, nil, nil}
	if conn := l.connect(); conn == nil {
		panic("Connection failed to database")
	}
	return l
}

func (l *LazyPdo) AddInitCommand(command string) {
	l.commands = append(l.commands, command)
}

// pdo connect to database
// should deals with compatible databases
func (l *LazyPdo) connect() *sql.DB {
	if l.pdo == nil {
		var err error
		var auth string
		splitDsn := strings.SplitN(l.dsn, ":", 2)
		dsn := splitDsn[1]
		switch splitDsn[0] {
		case "mysql":
			//user:password@tcp(127.0.0.1:3306)/database
			if l.user != "" && l.password != "" {
				auth = fmt.Sprintf("%s:%s@", l.user, l.password)
			}
			if l.pdo, err = sql.Open("mysql", fmt.Sprintf("%s%s", auth, dsn)); err != nil {
				log.Fatalf("Connection failed to database %s with error : %s", dsn, err)
			} else {
				log.Printf("Connected to %s", dsn)
			}
		case "pgsql":
			if l.user != "" && l.password != "" {
				auth = fmt.Sprintf(" user=%s password=%s ", l.user, l.password)
			}
			//Should add an option for ssl
			if l.pdo, err = sql.Open("postgres", fmt.Sprintf("%s %s sslmode=disable", auth, dsn)); err != nil {
				log.Fatalf("Connection failed to database %s with error : %s", dsn, err)
			} else {
				log.Printf("Connected to %s", dsn)
			}
		case "sqlsrv":
			if l.user != "" && l.password != "" {
				auth = fmt.Sprintf(";user id=%s;password=%s ", l.user, l.password)
			}
			if l.pdo, err = sql.Open("sqlserver", fmt.Sprintf("%s%s", dsn, auth)); err != nil {
				log.Fatalf("Connection failed to database %s with error : %s", dsn, err)
				return nil
			} else {
				log.Printf("Connected to %s", dsn)
			}
		case "sqlite":
			//file:test.s3db?_auth&_auth_user=admin&_auth_pass=admin
			if l.user != "" && l.password != "" {
				auth = fmt.Sprintf("&_auth&_auth_user=%s&_auth_pass=%s", l.user, l.password)
			}
			if l.pdo, err = sql.Open("sqlite3", fmt.Sprintf("%s%s", dsn, auth)); err != nil {
				log.Fatalf("Connection failed to database %s with error : %s", dsn, err)
				return nil
			} else {
				log.Printf("Connected to %s", dsn)
			}
		default:
		}
		for _, command := range l.commands {
			l.Query(command, "")
		}
	}
	return l.pdo
}

func (l *LazyPdo) Reconstruct(dsn string, user string, password string, options map[string]string) bool {
	l.dsn = dsn
	l.user = user
	l.password = password
	l.options = options
	l.commands = []string{}
	if l.pdo != nil {
		l.pdo = nil
		return true
	}
	return false
}

// To be done
func (l *LazyPdo) InTransaction() bool {
	// Do not call parent method if there is no pdo object
	//return $this->pdo && parent::inTransaction();
	return false
}

func (l *LazyPdo) SetAttribute(attribute, value interface{}) bool {
	/*if l.pdo != nil {
		l.connect().Attrs(attribute, value)
	}*/
	l.options[attribute.(string)] = value.(string)
	return true
}

func (l *LazyPdo) GetAttribute(attribute string) interface{} {
	/*if value, err := l.connect().Get(attribute); !err {
		return value
	}*/
	return nil
}

func (l *LazyPdo) BeginTransaction() (*sql.Tx, error) {
	return l.connect().BeginTx(context.Background(), nil)
}

// Should check return status
func (l *LazyPdo) Commit(tx *sql.Tx) error {
	return tx.Commit()
}

// Should check return status
func (l *LazyPdo) RollBack(tx *sql.Tx) error {
	return tx.Rollback()
}

/*
func (l *LazyPdo) ErrorCode(): mixed
{
	return $this->pdo()->errorCode();
}

func (l *LazyPdo) ErrorInfo(): array
{
	return $this->pdo()->errorInfo();
}

func (l *LazyPdo) Exec($query): int
{
	return $this->pdo()->exec($query);
}

func (l *LazyPdo) Prepare($statement, $options = array())
{
	return $this->pdo()->prepare($statement, $options);
}

func (l *LazyPdo) Quote($string, $parameter_type = null): string
{
	return $this->pdo()->quote($string, $parameter_type);
}

func (l *LazyPdo) LastInsertId($name = null): string
{
	return $this->pdo()->lastInsertId($name);
}
*/

func (l *LazyPdo) Query(sql string, parameters ...interface{}) ([]map[string]interface{}, error) {
	rows, err := l.connect().Query(sql, parameters...)
	if err != nil {
		return nil, err
	}
	results, err := l.Rows2Map(rows)
	return results, err
}

func (l *LazyPdo) QueryRowSingleColumn(sql string, parameters ...interface{}) (interface{}, error) {
	row := l.connect().QueryRow(sql, parameters...)
	var result interface{}
	if err := row.Scan(&result); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

// from https://kylewbanks.com/blog/query-result-to-map-in-golang
func (l *LazyPdo) Rows2Map(rows *sql.Rows) ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}
	cols, err := rows.Columns()
	if err != nil {
		return result, err
	}
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return result, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			switch v := (*val).(type) {
			case string:
				m[colName] = v
			case []uint8:
				m[colName] = string(v)
			default:
				m[colName] = *val
			}

		}

		result = append(result, m)
	}
	return result, err
}
