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

# creating the test files.
cd _SharedFiles
for i in `seq 1 8`;
do
  # If this is modified the test wont work bc its not hte same hash..
  printf "This is test$i" > test$i

done
cd ..


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
sleep 20 
# First A and B get some files 
# Below are hashes. 
# file1 = bcdf5d93e9c21cd36252731e2609e6c6a4af0b3634e612b19f97b2f17f09a1bb

# file2 = e40a0a31c20907a817cce838f7cc021933f0c7f9adff4c2b1de4053d7e8c4060
# file3 =fc1d7a4464c45001910c1ea75fd04a968a7c8fd3bf9d6069932475d3a5eb19d4
# file4 = c18cfb8c3dc98ab65a7ab8f03736b76e4477476010fc7ef03b58fa27c5d54159
# file5 = 70faeddff69d57e87c4c61d43d6ff167c55b565015f00112e41d3d72bf5d5b72
# file6 =95f211ae027ccf7aa136ca7b1e7a8b150ac9dc096ef59b9cf07e9d4aa8e427af
# file7 =8ec6833586971df62bd055f8c260f292562d3fdd88d9cb841b38af4c34d8c153
# file8 =a3ce8a16904f0bd29855edeb10d70886426ee929f1edd1e0210e449018aaf6bb
# main.go = 6b25fba702e9f5292a096e0d85e445f646a03807afaa84b71a19371003fd8d4a

./client/client -UIPort=12345 -file=test1
./client/client -UIPort=12345 -file=test2
./client/client -UIPort=12345 -file=test3
./client/client -UIPort=12345 -file=test4
./client/client -UIPort=12345 -file=test5
./client/client -UIPort=12345 -file=test6
./client/client -UIPort=12345 -file=test7
./client/client -UIPort=12345 -file=test8


# Then H and C download some
echo -e "First part of dl going on"
sleep 4
./client/client -UIPort=12352 -file=test1AH -dest=A -request=bcdf5d93e9c21cd36252731e2609e6c6a4af0b3634e612b19f97b2f17f09a1bb
./client/client -UIPort=12352 -file=test2AH -dest=A -request=e40a0a31c20907a817cce838f7cc021933f0c7f9adff4c2b1de4053d7e8c4060
./client/client -UIPort=12352 -file=test3AH -dest=A -request=fc1d7a4464c45001910c1ea75fd04a968a7c8fd3bf9d6069932475d3a5eb19d4
./client/client -UIPort=12352 -file=test4AH -dest=A -request=c18cfb8c3dc98ab65a7ab8f03736b76e4477476010fc7ef03b58fa27c5d54159


./client/client -UIPort=12347 -file=test5AC -dest=A -request=70faeddff69d57e87c4c61d43d6ff167c55b565015f00112e41d3d72bf5d5b72
./client/client -UIPort=12347 -file=test6AC -dest=A -request=95f211ae027ccf7aa136ca7b1e7a8b150ac9dc096ef59b9cf07e9d4aa8e427af
./client/client -UIPort=12347 -file=test7AC -dest=A -request=8ec6833586971df62bd055f8c260f292562d3fdd88d9cb841b38af4c34d8c153
./client/client -UIPort=12347 -file=test8AC -dest=A -request=a3ce8a16904f0bd29855edeb10d70886426ee929f1edd1e0210e449018aaf6bb


# Finally C and D download but from H and C
sleep 10

echo -e "Second part of DL going on "
./client/client -UIPort=12347 -file=test1HC -dest=H -request=bcdf5d93e9c21cd36252731e2609e6c6a4af0b3634e612b19f97b2f17f09a1bb
./client/client -UIPort=12347 -file=test2HC -dest=H -request=e40a0a31c20907a817cce838f7cc021933f0c7f9adff4c2b1de4053d7e8c4060
./client/client -UIPort=12347 -file=test3HC -dest=H -request=fc1d7a4464c45001910c1ea75fd04a968a7c8fd3bf9d6069932475d3a5eb19d4
./client/client -UIPort=12347 -file=test4HC -dest=H -request=c18cfb8c3dc98ab65a7ab8f03736b76e4477476010fc7ef03b58fa27c5d54159

./client/client -UIPort=12348 -file=test5CD -dest=C -request=70faeddff69d57e87c4c61d43d6ff167c55b565015f00112e41d3d72bf5d5b72
./client/client -UIPort=12348 -file=test6CD -dest=C -request=95f211ae027ccf7aa136ca7b1e7a8b150ac9dc096ef59b9cf07e9d4aa8e427af
./client/client -UIPort=12348 -file=test7CD -dest=C -request=8ec6833586971df62bd055f8c260f292562d3fdd88d9cb841b38af4c34d8c153
./client/client -UIPort=12348 -file=test8CD -dest=C -request=a3ce8a16904f0bd29855edeb10d70886426ee929f1edd1e0210e449018aaf6bb


#./client/client -UIPort=12348 -file=maingoDG.go -dest=G -request=6b25fba702e9f5292a096e0d85e445f646a03807afaa84b71a19371003fd8d4a
sleep 10
pkill -f Peerster




# cheack manually if it works in the _Download folder.
