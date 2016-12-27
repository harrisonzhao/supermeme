
#!/bin/bash

DBUSER=superanswer
DBPASS=supermeme2
DBHOST=52.186.123.148
DBNAME=alpha

DB=mysql://$DBUSER:$DBPASS@$DBHOST/$DBNAME

DEST=$1
if [ -z $DEST ]; then
	echo need to provide models_directory, usage is ./gen.sh models_directory
	exit
fi

if [ ! -d $DEST ]; then
	echo $DEST directory does not exist
	exit
fi

rm -f $DEST/*.go

XOBIN=$(which xo)

$XOBIN $DB -o $DEST
