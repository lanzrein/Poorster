echo -e "Starting a ring expansive test"


RED='\033[0;31m'
NC='\033[0m'
DEBUG="false"


UIPort=12345
gossipPort=5000
name='A'

# General peerster (gossiper) command
#./Peerster -UIPort=12345 -gossipPort=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 > A.out &
cd ..
go build

for i in `seq 1 10`;
do
	outFileName="$name.out"
	peerPort=$((($gossipPort+1)%10+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -rtimer=10 -peers=$peer > $outFileName &
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort"

	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done

echo -e "all peerster started..waiting for routing.."

cd client
go build

sleep 15

./client -UIPort=12345 -file=test1
./client -UIPort=12346 -file=test2


sleep 2

echo -e "A and B have resp. test1 and test2"



echo -e "Starting filesearch from C should find only test2"


./client -UIPort=12347 -keywords=test -budget=3

sleep 2


echo -e "Starting filesearch from F with ring expanding scheme. Should find all test"

./client -UIPort=12350 -keywords=test


sleep 5


echo -e "Downloading implicitly test1 and test2 for F"

./client -UIPort=12350 -request=bcdf5d93e9c21cd36252731e2609e6c6a4af0b3634e612b19f97b2f17f09a1bb -file=test1f
./client -UIPort=12350 -request=e40a0a31c20907a817cce838f7cc021933f0c7f9adff4c2b1de4053d7e8c4060 -file=test2f


sleep 5

echo -e "Searching from G Should find both file at F"

./client -UIPort=12351 -keywords=test -budget=3

sleep 5

pkill -f Peerster