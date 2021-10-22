package database

import (
	"log"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// LazyPdo is a custom db client
type LazyPdo struct {
	dsn      string
	user     string
	password string
	options  map[string]string
	commands []string
	pdo      *gorm.DB
}

func NewLazyPdo(dsn string, user string, password string, options map[string]string) *LazyPdo {
	return &LazyPdo{dsn, user, password, options, nil, nil}
}

func (l *LazyPdo) AddInitCommand(command string) {
	l.commands = append(l.commands, command)
}

// pdo connect to database
// should deals with compatible databases
func (l *LazyPdo) connect() *gorm.DB {
	if l.pdo == nil {
		splitDsn := strings.Split(l.dsn, ":")
		switch splitDsn[0] {
		case "mysql":
		case "pgsql":
		case "sqlsrv":
		case "sqlite":
			var err error
			if l.pdo, err = gorm.Open(sqlite.Open(splitDsn[1]), &gorm.Config{}); err != nil {
				log.Fatalf("Connection failed to database %s", splitDsn[1])
				return nil
			} else {
				log.Printf("Connected to %s", splitDsn[1])
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
	if l.pdo != nil {
		l.connect().Attrs(attribute, value)
	}
	l.options[attribute.(string)] = value.(string)
	return true
}

func (l *LazyPdo) GetAttribute(attribute string) interface{} {
	if value, err := l.connect().Get(attribute); !err {
		return value
	}
	return nil
}

func (l *LazyPdo) BeginTransaction() bool {
	l.connect().Begin()
	return true
}

// To be removed
func (l *LazyPdo) PDO() *gorm.DB {
	l.pdo = l.connect()
	return l.pdo
}

/*
func (l *LazyPdo) Commit(): bool
{
	return $this->pdo()->commit();
}

func (l *LazyPdo) RollBack(): bool
{
	return $this->pdo()->rollBack();
}

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
func (l *LazyPdo) Query(query string, fetchMode string, fetchModeArgs ...interface{}) []map[string]interface{} {
	// fetchMode useful ?
	var results []map[string]interface{}
	l.pdo.Raw(query, fetchModeArgs...).Scan(&results)
	return results
}

/*
func (l *LazyPdo) Connect() error {
	var err error
	if c.Client, err = gorm.Open(sqlite.Open("../../test/test.db"), &gorm.Config{}); err != nil {
		return err
	} else {
		log.Printf("Connected to %s", c.Client.Config.Name())
		return nil
	}
}*/
