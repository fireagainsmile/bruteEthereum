# bruteEthereum
Brute generating ethereum private keys and check them against known addresses.

According to some references, the actual probability of finding a private key in use is about 30000000 / 115792089237316195423570985008687907853269984665640564039457584007913129639936. 
Even at millions of privatekeys a second, it will be a very long time and huge amount of resources before you will likely find anything. 
Though, by leaving it running do have the possibility bring you some fortune.

# How does it work
A redis client is needed.

By configuring the provider in the `config.ini` and start the program, the program will subscribe the new heads event and get the latest blocks to save related transaction addresses to redis, as it is the lastest block, the block we download could be a canonical block or a uncle block or isolated block in the future. Meanwhile two worker brute forcely generating key pair to match know address without a break.

if this program have not running for a long time, it will cost some time to update blocks to last running state or untill arise an nil block.

Two hash worker running all the time using different algrithm. one generate private key randomly, one not. both of them will regenerate a new number as a new scanning point. It has the possibility that the two workers scan the same point, it also increases the possibility to find something.