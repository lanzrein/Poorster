cd ..
go build
./Peerster -name=Aleksandar -GUIPort=8000 -UIPort=8080 -peers=127.0.0.1:5001 -rtimer=10 -N=4 -stubbornTimeout=5  > A.out &
./Peerster -name=Bob -GUIPort=8001 -UIPort=8081 -gossipAddr=127.0.0.1:5001 -peers=127.0.0.1:5002 -rtimer=10 -N=4 -stubbornTimeout=5 > B.out &
./Peerster -name=Cynthia -GUIPort=8002 -UIPort=8082 -gossipAddr=127.0.0.1:5002 -peers=127.0.0.1:5003 -rtimer=10 -N=4 -stubbornTimeout=5  > C.out &
./Peerster -name=Daniel -GUIPort=8003 -UIPort=8083 -gossipAddr=127.0.0.1:5003 -peers=127.0.0.1:5000 -rtimer=10 -N=4 -stubbornTimeout=5  > C.out &

sleep 1000
