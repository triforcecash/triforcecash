# Usage
## Command line args
### -seed 
Password for account, anyone who knows them can dispose of account.
It should be random and long so that it cannot be cracked.
The program does not store it, if you lose the seed, you will lose access to the account.
### -port
Port (default 8075) 
### -host
Your public ip

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
