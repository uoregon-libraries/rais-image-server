ver=$(grep Version ./src/version/version.go | sed 's/^.*"\(.*\)".*$/\1/')
docker tag uolibraries/rais:prod uolibraries/rais:$ver
