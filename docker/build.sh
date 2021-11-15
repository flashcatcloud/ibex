#!/bin/sh

if [ $# -ne 1 ]; then
	echo "$0 <tag>"
	exit 0
fi

tag=$1

echo "tag: ${tag}"

rm -rf ibex && cp ../ibex . && docker build -t ibex:${tag} .

docker tag ibex:${tag} ulric2019/ibex:${tag}
docker push ulric2019/ibex:${tag}

