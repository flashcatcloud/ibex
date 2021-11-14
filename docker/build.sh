#!/bin/sh

if [ $# -ne 1 ]; then
	echo "$0 <tag>"
	exit 0
fi

tag=$1

echo "tag: ${tag}"

# clean workspace
rm -rf ibex.tar.gz ibex etc

# handle binary
cp ../ibex . && chmod +x ibex

# handle configuration
cp -r ../etc .
sed -i 's/127.0.0.1:3306/mysql:3306/g' ./etc/server.conf
sed -i 's/127.0.0.1:5432/postgres:5432/g' ./etc/server.conf
sed -i 's/127.0.0.1:20090/ibex:20090/g' ./etc/agentd.conf

# make tarball and delete tmp files
tar zcvf ibex.tar.gz ibex etc && rm -rf ibex etc

docker build -t ibex:${tag} .
