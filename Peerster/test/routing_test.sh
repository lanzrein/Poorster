cd ..
go build
cd client
go build
cd ..

echo -e "Setting UP A-B-C-D"
./Peerster -UIPort=12345 -gossipAddr=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 -rtimer=10 > A.out &
./Peerster -UIPort=12346 -gossipAddr=127.0.0.1:5002 -name=B -peers=127.0.0.1:5001 -rtimer=10 > B.out &
./Peerster -UIPort=12347 -gossipAddr=127.0.0.1:5003 -name=C -peers=127.0.0.1:5002 -rtimer=10 > C.out &
./Peerster -UIPort=12348 -gossipAddr=127.0.0.1:5004 -name=D -peers=127.0.0.1:5003 -rtimer=10 > D.out &


go build

./client/client -UIPort=12345 -msg="HI FROM A"
./client/client -UIPort=12346 -msg="HI FROM B"
./client/client -UIPort=12347 -msg="HI FROM C"
./client/client -UIPort=12348 -msg="HI FROM D"

sleep 10


./Peerster -UIPort=12349 -gossipAddr=127.0.0.1:5005 -name=E -peers=127.0.0.1:5001,127.0.0.1:5004 -rtimer=10 > E.out &
sleep 5

echo -e "E has been added to the network..."



./client/client -UIPort=12345 -msg="HI FROM A again !"

sleep 20
pkill -f Peerster
echo -e "Done "


# At this point D should discover the shortcut to A..


