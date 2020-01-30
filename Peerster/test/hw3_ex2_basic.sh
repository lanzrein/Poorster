cd ..
go build

./Peerster -name=A -UIPort=8080 -peers=127.0.0.1:5001 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true > A.out &
./Peerster -name=B -UIPort=8081 -peers=127.0.0.1:5000,127.0.0.1:5002 -gossipAddr=127.0.0.1:5001 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true  > B.out &
./Peerster -name=C -UIPort=8082 -peers=127.0.0.1:5001 -gossipAddr=127.0.0.1:5002 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true > C.out &


sleep 10

cd client
go build

./client -UIPort=8081 -file=test1
./client -UIPort=8082 -file=test2
./client -UIPort=8081 -file=test3

sleep 10

echo -e "Now we add a new node that will have to get all the TLC and gossiping things..."

cd ..


./Peerster -name=D -UIPort=8083 -peers=127.0.0.1:5000 -gossipAddr=127.0.0.1:5003 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true > D.out &

sleep 10


pkill -f Peerster
