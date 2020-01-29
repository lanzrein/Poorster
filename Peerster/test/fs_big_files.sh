


cd ..
# cleaning up the downloads.

cd _Downloads
rm *
cd ..

# building
go build
cd client
go build
cd ..

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
DEBUG="true"

outputFiles=()
make clean


UIPort=12345
gossipPort=5000
name='A'

# General peerster (gossiper) command
#./Peerster -UIPort=12345 -gossipAddr=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 > A.out &


for i in `seq 1 10`;
do
	outFileName="$name.out"
	peerPort=$((($gossipPort+1)%10+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer -rtimer=10 > $outFileName &
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort"
	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done
# Wait for routing.
./client/client -UIPort=12345 -file=random_maxsize

echo -e "Establishing routes "
sleep 10
# First A and B get some files
# capital.pdf sha = 5005ef6a4fc0f2ba8ee4330532bafb765e4e1a2cb20da39b835b9b7895b06e3f


# Then H and C download some
echo -e "First part of dl going on"

./client/client -UIPort=12346 -file=randAB -dest=A -request=f052e7bf173c881fd0900fc54e21f24b862ecb29e74a6460a2bf38c45f68a3f6
#./client/client -UIPort=12347 -file=randAC -dest=A -request=f052e7bf173c881fd0900fc54e21f24b862ecb29e74a6460a2bf38c45f68a3f6
#./client/client -UIPort=12348 -file=randAD -dest=A -request=f052e7bf173c881fd0900fc54e21f24b862ecb29e74a6460a2bf38c45f68a3f6
#./client/client -UIPort=12349 -file=randAE -dest=A -request=f052e7bf173c881fd0900fc54e21f24b862ecb29e74a6460a2bf38c45f68a3f6



sleep 5

echo -e "Concurrent download from B "
./client/client -UIPort=12350 -file=randBF -dest=B -request=f052e7bf173c881fd0900fc54e21f24b862ecb29e74a6460a2bf38c45f68a3f6
#./client/client -UIPort=12351 -file=randDG -dest=D -request=f052e7bf173c881fd0900fc54e21f24b862ecb29e74a6460a2bf38c45f68a3f6
#./client/client -UIPort=12352 -file=randEH -dest=E -request=f052e7bf173c881fd0900fc54e21f24b862ecb29e74a6460a2bf38c45f68a3f6

sleep 10
pkill -f Peerster




# cheack manually if it works in the _Download folder.