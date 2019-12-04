cd ..
go build

./Peerster -name=A -UIPort=8080 -peers=127.0.0.1:5001 -rtimer=10 > A.out &
sleep 1
./Peerster -name=B -UIPort=8081 -peers=127.0.0.1:5000,127.0.0.1:5002 -gossipAddr=127.0.0.1:5001 -rtimer=10 > B.out &
./Peerster -name=C -UIPort=8082 -peers=127.0.0.1:5001 -gossipAddr=127.0.0.1:5002 -rtimer=10 > C.out &

sleep 2

cd client
go build
./client -UIPort=8081 -file=test1
./client -UIPort=8082 -file=test1


sleep 5

echo -e "Starting a file search from A - Should find both"
./client -UIPort=8080 -keywords=test -budget=3


sleep 2

echo -e "Trying to download searched file.."
./client -UIPort=8080 -request=bcdf5d93e9c21cd36252731e2609e6c6a4af0b3634e612b19f97b2f17f09a1bb -file=search1

sleep 10

pkill -f Peerster