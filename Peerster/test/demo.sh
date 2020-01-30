echo -e "Starting script for demo..."

cd ..
go build
xterm -T "PeersterA" -n "PeertserA" -e ./Peerster -name=A -GUIPort=8000 -UIPort=8080 -peers=127.0.0.1:5001 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true &
xterm -T "PeersterB" -n "PeertserB" -e ./Peerster -name=B -GUIPort=8001 -UIPort=8081 -gossipAddr=127.0.0.1:5001 -peers=127.0.0.1:5002 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true &
xterm -T "PeersterC" -n "PeertserC" -e ./Peerster -name=C -GUIPort=8002 -UIPort=8082 -gossipAddr=127.0.0.1:5002 -peers=127.0.0.1:5003 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true &
xterm -T "PeersterD" -n "PeertserD" -e ./Peerster -name=D -GUIPort=8003 -UIPort=8083 -gossipAddr=127.0.0.1:5003 -peers=127.0.0.1:5000 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true &

sleep 10

cd client

go build
./client -UIPort=8080 -msg="A"
./client -UIPort=8081 -msg="B"
./client -UIPort=8082 -msg="C"
./client -UIPort=8083 -msg="D"
sleep 2

./client -UIPort=8080 -initcluster
./client -UIPort=8081 -joinOther=A
sleep 1
./client -UIPort=8080 -accept=B
sleep 1
./client -UIPort=8082 -joinOther=A
sleep 1
./client -UIPort=8080 -accept=C
./client -UIPort=8081 -deny=C

sleep 1
./client -UIPort=8083 -joinOther=A
sleep 2

./client -UIPort=8080 -accept=D
./client -UIPort=8081 -accept=D
./client -UIPort=8082 -accept=D


sleep 1

./client -UIPort=8080 -broadcast -msg="Hello friends!"
