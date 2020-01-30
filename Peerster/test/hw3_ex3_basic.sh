cd ..
go build

./Peerster -name=A -UIPort=8080 -peers=127.0.0.1:5001 -rtimer=10 -N=3 -stubbornTimeout=5 -hw3ex3=true > A.out &
./Peerster -name=B -UIPort=8081 -peers=127.0.0.1:5000,127.0.0.1:5002 -gossipAddr=127.0.0.1:5001 -rtimer=10 -N=3 -stubbornTimeout=5 -hw3ex3=true  > B.out &
./Peerster -name=C -UIPort=8082 -peers=127.0.0.1:5001 -gossipAddr=127.0.0.1:5002 -rtimer=10 -N=3 -stubbornTimeout=5 -hw3ex3=true > C.out &


sleep 10

cd client
go build

echo -e "Trying to have two process advance to next round \n A should NOT advance to next round"
./client -UIPort=8081 -file=test1
./client -UIPort=8082 -file=test2
#./client -UIPort=8081 -file=test3

sleep 10

echo -e "Now A will try to add a file and fail - then move to next round"

./client -UIPort=8080 -file=test3

sleep 10

pkill -f Peerster