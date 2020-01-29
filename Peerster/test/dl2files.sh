cd ..

./client/client -file=test1
./client/client -file=test2


sleep 3


./client/client -UIPort=8081 -file=test1Res -dest=A -request=bcdf5d93e9c21cd36252731e2609e6c6a4af0b3634e612b19f97b2f17f09a1bb
./client/client -UIPort=8081 -file=test2res -dest=A -request=e40a0a31c20907a817cce838f7cc021933f0c7f9adff4c2b1de4053d7e8c4060