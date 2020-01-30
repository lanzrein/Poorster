echo -e "Starting script for demo..."

cd ..
go build
xterm -T "PeersterA" -n "PeertserA" -e ./Peerster -name=Alex -GUIPort=8000 -UIPort=8080 -peers=127.0.0.1:5001 -rtimer=10 -N=4 -stubbornTimeout=5  &
xterm -T "PeersterB" -n "PeertserB" -e ./Peerster -name=Bob -GUIPort=8001 -UIPort=8081 -gossipAddr=127.0.0.1:5001 -peers=127.0.0.1:5002 -rtimer=10 -N=4 -stubbornTimeout=5 &
xterm -T "PeersterC" -n "PeertserC" -e ./Peerster -name=Cynthia -GUIPort=8002 -UIPort=8082 -gossipAddr=127.0.0.1:5002 -peers=127.0.0.1:5003 -rtimer=10 -N=4 -stubbornTimeout=5 &
xterm -T "PeersterD" -n "PeertserD" -e ./Peerster -name=Diana -GUIPort=8003 -UIPort=8083 -gossipAddr=127.0.0.1:5003 -peers=127.0.0.1:5000 -rtimer=10 -N=4 -stubbornTimeout=5 &

sleep 10

cd client

go build
./client -UIPort=8080 -msg="bob"
./client -UIPort=8081 -msg="bob"
./client -UIPort=8082 -msg="bob"
./client -UIPort=8083 -msg="bob"
sleep 2

./client -UIPort=8080 -initcluster
./client -UIPort=8081 -joinOther=Alex
sleep 2
./client -UIPort=8080 -accept=Bob
sleep 3
./client -UIPort=8082 -joinOther=Alex
sleep 3
./client -UIPort=8080 -accept=Cynthia
./client -UIPort=8081 -deny=Cynthia


./client -UIPort=8083 -joinOther=Alex
sleep 3
./client -UIPort=8080 -accept=Diana
./client -UIPort=8081 -deny=Diana
./client -UIPort=8082 -accept=Diana

sleep 30

./client -UIPort=8080 -expelOther=Bob
sleep 3
./client -UIPort=8080 -accept=Bob
./client -UIPort=8082 -deny=Bob
./client -UIPort=8083 -accept=Bob


sleep 1

./client -UIPort=8080 -broadcast -msg="Hello friends!"

