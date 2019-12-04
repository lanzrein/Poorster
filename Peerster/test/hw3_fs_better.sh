cd ..
go build

./Peerster -name=A -UIPort=8080 -peers=127.0.0.1:5001 -rtimer=10 > A.out &
sleep 1
./Peerster -name=B -UIPort=8081 -peers=127.0.0.1:5000 -gossipAddr=127.0.0.1:5001 -rtimer=10 > B.out &
./Peerster -name=C -UIPort=8082 -peers=127.0.0.1:5000 -gossipAddr=127.0.0.1:5002 -rtimer=10 > C.out &
./Peerster -name=D -UIPort=8083 -peers=127.0.0.1:5000 -gossipAddr=127.0.0.1:5003 -rtimer=10 > D.out &

sleep 5

cd client
go build
./client -UIPort=8081 -file=test1
./client -UIPort=8082 -file=test1


sleep 5

echo -e "Starting a file search from A - Should find nothing but go all the way. - also doing budget a lot."

./client -UIPort=8080 -keywords=test -budget=5
sleep 5

./client -UIPort=8080 -keywords=blibli
./client -UIPort=8080 -keywords=ba -budget=0

sleep 6

./client -UIPort=8080 -keywords=bla -budget=1


echo -e "Trying to download searched file.."
./client -UIPort=8080 -request=bcdf5d93e9c21cd36252731e2609e6c6a4af0b3634e612b19f97b2f17f09a1bb -file=search1

sleep 10

pkill -f Peerster