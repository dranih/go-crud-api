#!/bin/bash

function usage {
    echo "Usage : start_db.sh <mysql|pgsql|sqlsrv|all> [clean]"
}

function clean {
    #Remove container if exists
    if [ $( docker ps -a -f name=${containerName} | wc -l ) -eq 2 ]; then
        echo "Deleting container ${containerName}"
        docker rm -f ${containerName}
    fi
}

function test {
    code=$1
    msg=$2
    if [ ${code} != 0 ]; then
        echo $msg
        exit $code
    fi
}

function started {
    echo "${dbType} database started in container ${containerName}"
    echo "GCA_CONFIG_FILE must be set before runnig test by running :"
    echo "export GCA_CONFIG_FILE=${configFile}"
}

function isContainerRunning {
    [ z$( docker ps -q -f name=${containerName} ) != z ]
}

function startAndLoad {
    case ${dbType} in
        mysql)
            dbPort=3306
            startContainerCmd="-e MYSQL_ALLOW_EMPTY_PASSWORD=true -d mysql:latest"
            createDbCmd="mysql -uroot < /gocrud/create_mysql.sql"
            loadSqlCmd="mysql go-crud-api -ugo-crud-api -pgo-crud-api < /gocrud/blog_mysql.sql"
            ;;
        pgsql)
            dbPort=5432
            startContainerCmd="-e POSTGRES_PASSWORD=go-crud-api -d postgis/postgis"
            createDbCmd="PGPASSWORD=go-crud-api psql -q -U postgres -f /gocrud/create_pgsql.sql"
            loadSqlCmd="PGPASSWORD=go-crud-api psql -q -U go-crud-api -f /gocrud/blog_pgsql.sql"
            ;;
        sqlsrv)
            dbPort=1433
            startContainerCmd="-e ACCEPT_EULA=Y -e SA_PASSWORD=mystrongPasw0rd! -d mcr.microsoft.com/mssql/server:2019-latest"
            createDbCmd="/opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P mystrongPasw0rd! -i /gocrud/create_sqlsrv.sql"
            loadSqlCmd="/opt/mssql-tools/bin/sqlcmd -S localhost -U go-crud-api -P go-crud-api -i /gocrud/blog_sqlsrv.sql"
            ;;
        *)
            echo "This should not happen"
            exit 254
            ;;
    esac

    echo ">>>>> ${dbType}"
    containerName="gocrud_${dbType}"
    configFile="${scriptDir}/yaml/gcaconfig_${dbType}.yaml"
    if [[ "${action}" != "clean" ]]; then
        if ! isContainerRunning ; then
            clean
            echo "Starting container ${containerName}"
            docker run -p ${dbPort}:${dbPort} --name ${containerName} -v ${scriptDir}/sql:/gocrud ${startContainerCmd}
            #Wait for db to start, should do it cleaner
            sleep 15
            docker exec ${containerName} bash -c "${createDbCmd}"
            test $? "Could not create database"
        else
            echo "Container ${containerName} already running"
        fi
        docker exec ${containerName} bash -c "${loadSqlCmd}"
        test $? "Could not load test sql in database"
        started
    else
        clean
    fi
}

# Start
if (! docker stats --no-stream > /dev/null 2>&1); then
    echo "Is docker running ?"
    exit 254
fi

dbType=$1
action=$2
scriptDir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

if [[ z"${dbType}" == z ]]; then
    usage
    exit 254
fi

case ${dbType} in
  mysql | pgsql | sqlsrv)
    startAndLoad
    ;;
  all)
    for dbType in mysql pgsql sqlsrv
    do
        startAndLoad
    done
    ;;
  *)
    usage
    ;;
esac
