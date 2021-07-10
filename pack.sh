go build main.go
rm -r -f pack-output
mkdir pack-output
cp ./main ./pack-output/youvideo
cp -a ./pack/. ./pack-output