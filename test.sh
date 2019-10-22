#!/bin/sh

echo "sleep for 10 seconds to let the istio sidecar come up"
sleep 10

echo "test postgres master and replica with URLs ${URL_MASTER} and ${URL_REPLICA}"

psql -h ${URL_MASTER} -U ${USERNAME} -c "create database test;"
if [ $? != 0 ]  
then
	echo "Error creating database test"
	exit 1
fi

echo "database created at ${URL_MASTER}"

psql -h ${URL_MASTER} -U ${USERNAME} -c "create table posts (
id integer,
title character varying(100),
content text, 
published_at timestamp without time zone,
type character varying(100)
);
"

if [ $? != 0 ]  
then
	echo "Error creating table"
	exit 1
fi

echo "table created at ${URL_MASTER}"

psql -h ${URL_MASTER} -U ${USERNAME} -c "
insert into posts (id, title, content, published_at, type) values
(100, 'Intro to SQL', 'Epic SQL Content', '2018-01-01', 'SQL'),
(101, 'Intro to PostgreSQL', 'PostgreSQL is awesome!', now(), 'PostgreSQL');
"


if [ $? != 0 ]  
then
	echo "Error inserting rows"
	exit 1
fi

echo "rows inserted at ${URL_MASTER}"

RES=$(psql -h ${URL_MASTER} -U ${USERNAME} -qat -c "select * from posts;" | wc -l)

if [ $? != 0 ]  
then
	echo "Error selecting rows"
	exit 1
fi

echo "rows found at ${URL_MASTER}"

if [ $RES != 4 ]  
then
	echo "Expected 4 Rows, got $RES"
	exit 1
fi

echo "expected number of rows found at ${URL_MASTER}"

RES=$(psql -h ${URL_REPLICA} -U ${USERNAME} -qat -c "select * from posts;" | wc -l)

if [ $? != 0 ]  
then
	echo "Error selecting rows"
	exit 1
fi

if [ $RES != 4 ]  
then
	echo "Expected 4 Rows, got $RES"
	exit 1
fi

echo "expected number of rows found at ${URL_REPLICA}"

psql -h ${URL_REPLICA} -U ${USERNAME} -c "delete from posts where id!=1;" 

if [ $? == 0 ]
then
        echo "Deleting succeeded on the read only replica, this is an error"
        exit 1
fi

echo "no insert possible in read replica at ${URL_REPLICA}  --> that's correct"

psql -h ${URL_MASTER} -U ${USERNAME} -c "delete from posts where id!=1;" 

if [ $? != 0 ]
then
        echo "Error deleting rows on the master"
        exit 1
fi

echo "deleted rows in ${URL_MASTER} "

psql -h ${URL_MASTER} -U ${USERNAME} -c "drop table posts" 

if [ $? != 0 ]
then
        echo "Error dropping table on the master"
        exit 1
fi

psql -h ${URL_MASTER} -U ${USERNAME} -c "drop database test" 

if [ $? != 0 ]
then
        echo "Error deleting database on the master"
        exit 1
fi

echo "success"

exit 0
