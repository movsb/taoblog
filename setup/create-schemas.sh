#!/bin/bash

echo "Input DB settings for taoblog: "

read -p "DB_HOST: " DB_HOST
read -p "DB_NAME: " DB_NAME
read -p "DB_USER: " DB_USER
read -p "DB_PASS: " DB_PASS

read -p "Input mysql username: " MYSQL_USER

query() {
    echo "BEGIN;"

    echo "CREATE DATABASE $DB_NAME;"
    echo "GRANT ALL PRIVILEGES ON $DB_NAME.* TO \"$DB_USER\"@\"$DB_HOST\" IDENTIFIED BY \"$DB_PASS\";"
    echo "FLUSH PRIVILEGES;"

    echo "USE $DB_NAME;"

    cat schemas.sql

    echo "COMMIT;"
}

query | mysql -u$MYSQL_USER -h "$DB_HOST" -p
