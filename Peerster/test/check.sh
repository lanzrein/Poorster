cd ..
go build
./Peerster -name=A -GUIPort=8000 -UIPort=8080 -peers=127.0.0.1:5001 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true > A.out &
./Peerster -name=B -GUIPort=8001 -UIPort=8081 -gossipAddr=127.0.0.1:5001 -peers=127.0.0.1:5002 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true > B.out &
./Peerster -name=C -GUIPort=8002 -UIPort=8082 -gossipAddr=127.0.0.1:5002 -peers=127.0.0.1:5003 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true > C.out &
#./Peerster -name=D -UIPort=8083 -gossipAddr=127.0.0.1:5003 -peers=127.0.0.1:5000 -rtimer=10 -N=4 -stubbornTimeout=5 -ackAll=true > /dev/null &

sleep 1000
