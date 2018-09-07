# Build
```bash
go get github.com/syndtr/goleveldb/leveldb
go get golang.org/x/crypto/sha3
go get golang.org/x/crypto/ed25519
go get github.com/andlabs/ui
go get github.com/triforcecash/triforcecash
cd $GOPATH/src/github.com/triforcecash/triforcecash
go build node.go
go build gui.go
```
# Usage
## Command line args
### -seed 
Password for account, anyone who knows them can dispose of account.
It should be random and long so that it cannot be cracked.
The program does not store it, if you lose the seed, you will lose access to the account.
### -port
Port (default 8075) 
### -host
Your public ip (required for mining).
### -lobby
Should specify if default lobby does not work.
## Example
### Run
```bash
nohup triforcecash_node_linux_amd64 -seed mypassword -host 123.123.123.123 & 
```
Port 8075 must be open for tcp connections.
To check in browser enter address your_ip:8075/api/hosts
### Stop
```bash
killall triforcecash_node_linux_amd64 
```
## CPU Miner
You can run an unlimited number of miners on different computers that will calculate the nonce for one node.
### Run
Copy public key from Receive tab in GUI Wallet.
Enter the node ip address and port after the -host flag.
```bash
 ./miner -host 127.0.0.1:8075 -publickey a363f3675039caf20b8f805479051482e3c87b69d39b9b94f568778e8335a586 -threads 6
```
## API

### Create Account

#### Python2.7:
```python
import json
import requests
myseed=b'passwordpasswordpasswordpassword'
jsn={"Seed":myseed.encode("base64")}
req=requests.post("http://127.0.0.1:8075/api/genaccount",json.dumps(jsn))
jsn=req.json()
jsn["Addr"].decode("base64").encode("hex")
#'2b560d9daefc215c84eebec91c47893c616df5f4ab615cdee6ae83437a091878'
jsn["Pub"].decode("base64").encode("hex")
#'6a771912acadc739b041b58f0cee218b8aa8b125b63b5bc850ed8726af1e4aea'
```


### Transfer

#### Python2.7:
```python
import json
import requests
myseed=b'passwordpasswordpasswordpassword'
addr=b'2b560d9daefc215c84eebec91c47893c616df5f4ab615cdee6ae83437a091878'.decode("hex")
amount=100
fee=1
jsn={"Seed":myseed.encode("base64"),"Addr":addr.encode("base64"),"Amount":amount,"Fee":fee}
requests.post("http://127.0.0.1:8075/api/send",json.dumps(jsn))
```
### Get account state
#### Python2.7
```python
import requests
addr=b'2b560d9daefc215c84eebec91c47893c616df5f4ab615cdee6ae83437a091878'
req=requests.get("http://127.0.0.1:8075/api/statejson?key="+addr)
req.json()
#{u'Nonce': 1, u'Balance': 99, u'Addr': u'K1YNna78IVyE7r7JHEeJPGFt9fSrYVze5q6DQ3oJGHg=', u'LastBlockId': 838}
