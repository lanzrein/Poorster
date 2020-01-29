cd ..
go build
cd client
go build
cd ..

echo -e "Setting UP Line network "
./Peerster -UIPort=12345 -gossipAddr=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 -rtimer=10 > A.out &
./Peerster -UIPort=12346 -gossipAddr=127.0.0.1:5002 -name=B -peers=127.0.0.1:5001 -rtimer=10 > B.out &
./Peerster -UIPort=12347 -gossipAddr=127.0.0.1:5003 -name=C -peers=127.0.0.1:5002 -rtimer=10 > C.out &
./Peerster -UIPort=12348 -gossipAddr=127.0.0.1:5004 -name=D -peers=127.0.0.1:5003 -rtimer=10 > D.out &
./Peerster -UIPort=12349 -gossipAddr=127.0.0.1:5005 -name=E -peers=127.0.0.1:5004 -rtimer=10 > E.out &
./Peerster -UIPort=12350 -gossipAddr=127.0.0.1:5006 -name=F -peers=127.0.0.1:5005 -rtimer=10 > F.out &
./Peerster -UIPort=12351 -gossipAddr=127.0.0.1:5007 -name=G -peers=127.0.0.1:5006 -rtimer=10 > G.out &
./Peerster -UIPort=12352 -gossipAddr=127.0.0.1:5008 -name=H -peers=127.0.0.1:5007 -rtimer=10 > H.out &
./Peerster -UIPort=12353 -gossipAddr=127.0.0.1:5009 -name=I -peers=127.0.0.1:5008 -rtimer=10 > I.out &
./Peerster -UIPort=12354 -gossipAddr=127.0.0.1:5010 -name=J -peers=127.0.0.1:5009 -rtimer=10 > J.out &

echo -e "Initial messages to pump the routing."
./client/client -UIPort=12345 -msg="HI FROM A"
./client/client -UIPort=12346 -msg="HI FROM B"
./client/client -UIPort=12347 -msg="HI FROM C"
./client/client -UIPort=12348 -msg="HI FROM D"

sleep 15


echo -e "SENDING FROM J TO A."



./client/client -UIPort=12354 -msg="HI FROM J :)" -dest=A

sleep 10
#pkill -f Peerster
echo -e "Done "


# At this point D should discover the shortcut to A..


