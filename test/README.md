# Testing
* The folder **sql** contains the sql files to create the test database for sqlite, mysql, pgsql and sqlserver. These are copies from php-crud-api repo test folder.
* The folder **yaml** contains the gocrudapi config files to test with each type of database
* The bash script **start_db.sh** <mysql|pgsql|sqlserver> starts a database with docker of the choosen type and execute the sql file to it. The GCA_CONFIG_FILE has to be set to the adequat yaml file before running the tests.

By default (if GCA_CONFIG_FILE env var is empty), the tests which need a database will load a sqlite database into **/tmp/gocrudtests.db**