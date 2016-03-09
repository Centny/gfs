#!/bin/bash
for file in ./*;
do
	echo upload $file
	gfs -u http://127.0.0.1:2010 $file "" $1
done
