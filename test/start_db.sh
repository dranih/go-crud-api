#!/bin/bash

function usage {
    echo "Usage : start_db.sh <mysql|pgsql|sqlserver> [clean]"
}

function clean {
    #Remove container if exists
    if [ $( docker ps -a -f name=${containerName} | wc -l ) -eq 2 ]; then
        docker rm -f ${containerName}
    fi
}

function test {
    code=$1
    msg=$2
    if [[ $? != 0 ]]; then
        echo $msg
        exit $code
    fi
}

function started {
    echo "${dbType} database started in container ${containerName}"
    echo "GCA_CONFIG_FILE must be set before runnig test by running :"
    echo "export GCA_CONFIG_FILE=${configFile}"
}

if (! docker stats --no-stream > /dev/null 2>&1); then
    echo "Is docker running ?"
    exit 254
fi

dbType=$1
clean=$2
scriptDir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

if [[ z"${dbType}" == z ]]; then
    usage
    exit 254
fi

containerName="gocrud_${dbType}"
configFile="${scriptDir}/yaml/gcaconfig_${dbType}.yaml"

case ${dbType} in
  mysql)
    clean
    if [[ "${clean}" != "clean" ]]; then
        docker run -p 3306:3306 --name ${containerName} -v ${scriptDir}/sql:/gocrud -e MYSQL_ALLOW_EMPTY_PASSWORD=true -d mysql:latest
        #Wait for db to start, should do it cleaner
        sleep 15
        docker exec ${containerName} bash -c "mysql -uroot < /gocrud/create_mysql.sql"
        test $? "Could not create database"
        docker exec ${containerName} bash -c "mysql go-crud-api -ugo-crud-api -pgo-crud-api < /gocrud/blog_mysql.sql"
        test $? "Could not load test sql in database"
        started
    fi
    ;;
  pgsql)
    ;;
  sqlserver)
    ;;
  *)
    usage
    ;;
esac
