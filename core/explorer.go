package core

import (
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
)


var templates,err = template.New("templates").Parse(Templates)

const (
	



	Templates = `
{{define "Head"}}
<!DOCTYPE html>
<html>
<head>
	<title>explorer triforcecash</title>
</head>
<body>
	<style type="text/css">
	body{
		background-color:#102026;
		color:#cfcfdf;
		font-family: Verdana, sans-serif;
		font-weight: lighter;
		line-height: 20px;
		font-size: 13px;
	}

	h1,h2,h3,h4,h5,a,p,input,textarea{
		font-family: Verdana, sans-serif;
		font-weight: lighter;
		line-height: 14px;
	}

    a{
    	outline: none;
    	text-decoration: none;
    	color:#cfcfdf;
    }
    .container{
    	width: 90%;
    	padding-left: 5%;
    	padding-top:1%;
    	min-width: 900px;
    }

    .lineitem{
    	display: inline-block;
    }
    textarea:focus,input:focus{
    	outline: none
    }
    .keyentry{
    	height: 16px;
    	width: 500px;
    	background-color: transparent;
    	font-size: 13px;
    	font-family: Verdana, sans-serif;
		font-weight: lighter;
		line-height: 35px;
		color:#cfcfdf;
    	border: none;
    	border-bottom: solid;
    	border-width: 1px;
    	border-color: #cfcfdf;
    	outline: none;

    
    }

    .button{
    	background-color: transparent;
    	font-size: 13px;
    	font-family: Verdana, sans-serif;
		font-weight: lighter;
		line-height: 35px;
		color:#cfcfdf;
    	border: none;
    	outline: none;
    }

</style>
<div class="container">
	<div><div class="lineitem" style="padding-right: 5%"><a href="/"><h2>explorer.triforcecash.com</h2></a></div>
	<div class="lineitem" ><form><input class="keyentry" name="key" type="text" placeholder="search by address or block hash"><input type="submit" class="button" name="GO" value="GO"></form></div>
	</div>
{{end}}

{{define "End"}}	
</div>
</body>
</html>
{{end}}

	{{define "Account"}}
	<div>
		<div><h3>Account</h3></div>
		<div class="lineitem" style="padding-right: 1%" align="right">
			<p>Address</p>
			<p>Balance</p>
			<p>Nonce</p>
			<p>Confirm</p>
		</div>
		<div class="lineitem" style="padding-right: 1%" align="left">
			<p>{{.Addr}}</p>
			<p>{{.Balance}}</p>
			<p>{{.Nonce}}</p>
			<p>{{.Confirm}}</p>
		</div>
	</div>
	{{end}}

	{{define "Block"}}

	<div>
		<div><h3>Block</h3></div>
		<div class="lineitem" style="padding-right: 1%" align="right">
			<p>Hash</p>
			<p>Previous</p>
			<p>ID</p>
			<p>Rate</p>
			<p>Miner1</p>
			<p>Miner2</p>
			<p>Reward</p>
			<p>Fee</p>
			<p>Txs</p>
			<p>Accounts</p>
		</div>

		<div class="lineitem">
			<p>{{.Hash}}</p>
			<p><a href="?key={{.Prev}}">{{.Prev}}</a></p>
			<p>{{.Id}}</p>
			<p>{{.Rate}}</p>
			<p><a href="?key={{.Miner1}}">{{.Miner1}}</a></p>
			<p><a a href="?key={{.Miner2}}">{{.Miner2}}</a></p>
			<p>{{.Reward}}</p>
			<p>{{.Fee}}</p>
			<p>{{.Txs}}</p>
			<p>{{.Accounts}}</p>
		</div>
	</div>
	{{end}}
	{{define "Transactions"}}
	<div>
	<h3>Transactions</h3>
	<div style="border-top: 1px solid #afafdf ;"></div>
	{{range .TxsList}}
	<div>
		<div class="lineitem" style="padding-right: 1%" align="right" >
			<p><a href="?key={{.Sender}}">{{.Sender}}</a></p>
			<p style="color: #aa0000;">-{{.Amount}}</p>
		</div>
		<div class="lineitem" style="align-self: top;" align="right">
		{{range .Outs}}

			<p><a href="?key={{.Addr}}">{{.Addr}}</a></p>
			<p style="color: #00aa00;">+{{.Amount}}</p>
		{{end}}
		</div>
	</div>
	<div style="border-top: 1px solid #afafdf ;"></div>
	{{end}}
	<div>
	{{end}}


{{define "AccountTmplt"}}
{{template "Head"}}
{{template "Account" .}}
{{template "Transactions" .}}
{{template "End"}}
{{end}}

{{define "BlockTmplt"}}
{{template "Head"}}
{{template "Block" .}}
{{template "Transactions" .}}
{{template "End"}}
{{end}}

`
)

func ExplorerServ(res http.ResponseWriter,req *http.Request){
	hexkey:=req.URL.Query().Get("key")
	key,_:=hex.DecodeString(hexkey)
	if err!=nil{
		log.Println(err)
	}


	head,_,_,_,_:=GetHeader(key)

	if len(key)<32{
		head=Main.Higher
	}

	if head!=nil{
		templates.ExecuteTemplate(res.(io.Writer),"BlockTmplt",BlockToTemplate(head))
		return
	}
	state:=GetBalance(string(key))
	if state!=nil{

		templates.ExecuteTemplate(res.(io.Writer),"AccountTmplt",StateToTemplate(state))
		return
	}

}


type StateTmplt struct{
	Addr string
	Balance uint64
	Nonce uint64
	Confirm uint64
	TxsList []*TxTmplt
}

type BlockTmplt struct{
	Prev string
	Hash string
	Id uint64
	Rate string
	Miner1 string
	Miner2 string
	Reward uint64
	Fee uint64
	Txs uint64
	Accounts uint64
	TxsList []*TxTmplt
}

func StateToTemplate(state *State)*StateTmplt{
	hist:=GetTxsHistory(string(state.Addr))
	
	return &StateTmplt{
		Addr:fmt.Sprintf("%x",state.Addr),
		Balance:state.Balance,
		Confirm:state.Confirm,
		Nonce:state.Nonce,
		TxsList:TxsToTemplate(TxsHistoryToTxsList(hist)),
	}
}


func BlockToTemplate(head *Header)*BlockTmplt{
	txs,_:=GetTxsList(head.Txs)
	sts,_:=GetState(head.State)
	return &BlockTmplt{
		Prev:fmt.Sprintf("%x",head.Prev),
		Hash: fmt.Sprintf("%x",head.Hash()),
		Id: head.Id,
		Rate: head.Rate().String(),
		Miner1: fmt.Sprintf("%x",Addr(head.Pubs[0])),
		Miner2: fmt.Sprintf("%x",Addr(head.Pubs[1])),
		Reward: Reward(head.Id+1),
		Fee: head.Fee,
		Txs: uint64(len(txs)),
		Accounts: uint64(sts.Len()),
		TxsList:TxsToTemplate(txs),
	}
}

type TxOutTmplt struct{
	Addr string
	Amount uint64
}

type TxTmplt struct{
	Sender string
	Amount uint64
	Nonce uint64
	Outs []*TxOutTmplt
}

func TxToTemplate(tx *Tx)*TxTmplt{
	txtmplt:=&TxTmplt{
		Sender:fmt.Sprintf("%x",tx.Sender()),
		Amount:tx.Amount(),
		Nonce:tx.Nonce,
		Outs:make([]*TxOutTmplt,len(tx.Outs)),
	}
	for i, o := range tx.Outs{
		txtmplt.Outs[i]=&TxOutTmplt{
			Addr: fmt.Sprintf("%x", out(o).getAddr()),
			Amount: out(o).getAmount(),
		}
	}
	return txtmplt
}

func TxsToTemplate(txs TxsList)[]*TxTmplt{
	txstmplt:=make([]*TxTmplt,len(txs))
	for i,tx:=range txs{
		txstmplt[i]=TxToTemplate(tx)
	}
	return txstmplt
}

func TxsHistoryToTxsList(history []SearchTxsResultItem)TxsList{
	res:=TxsList{}
	for _,item:=range history{
		l:=len(item.TxsList)
		reverse:=make(TxsList,l)
		for i:=l-1;i>=0;i--{
			reverse[i]=item.TxsList[l-i-1]
		}
		res=append(res,reverse...)
	}
	return res
}
