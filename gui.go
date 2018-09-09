package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/andlabs/ui"
	"github.com/triforcecash/triforcecash/core"
	"time"
)

var (
	seed, pub, priv, addr []byte
	balance, conf         uint64
)

func main() {
	port := flag.Int("port", 8075, "Port")
	checkdepth := flag.Int("checkdepth", 1000, "For a stronger check, you should set 10000.")
	hostname := flag.String("host", "127.0.0.1", "Public ip")
	fullnode := flag.Bool("fullnode", false, "Will be able fullnode features")
	lobby := flag.String("lobby", "185.234.15.72:8075", "Lobby node")
	flag.Parse()
	core.FullNode = *fullnode
	core.Checkdepth = *checkdepth
	core.Port = fmt.Sprint(":", *port)
	core.PublicIp = *hostname
	if *hostname == "127.0.0.1" {
		core.ClientOnly = true
	}
	core.AddHostAddr(*lobby)
	core.Start()
	defer core.Stop()

	ui.Main(func() {
		Receive := ui.NewVerticalBox()
		Receive.SetPadded(true)
		Receive.Append(ui.NewHorizontalSeparator(), false)

		Receive.Append(ui.NewLabel("Your address:"), false)
		addrbox := ui.NewEntry()
		addrbox.SetReadOnly(true)
		Receive.Append(addrbox, false)

		Receive.Append(ui.NewLabel("Seed:"), false)
		seedbox := ui.NewEntry()
		Receive.Append(seedbox, false)

		genacc := ui.NewButton("Generate Account")
		hide := ui.NewButton("Hide")

		butbox0 := ui.NewHorizontalBox()
		butbox0.SetPadded(true)
		butbox0.Append(genacc, false)
		butbox0.Append(hide, false)
		Receive.Append(butbox0, false)

		Receive.Append(ui.NewLabel("Public Key:"), false)
		pubbox := ui.NewEntry()
		pubbox.SetReadOnly(true)
		Receive.Append(pubbox, false)
		Receive.Append(ui.NewLabel("Private Key:"), false)
		privbox := ui.NewEntry()
		privbox.SetReadOnly(true)
		Receive.Append(privbox, false)
		Receive.Append(ui.NewHorizontalSeparator(), false)

		Send := ui.NewVerticalBox()
		Send.SetPadded(true)
		Send.Append(ui.NewHorizontalSeparator(), false)
		Send.Append(ui.NewLabel("Pay to:"), false)
		paytobox := ui.NewEntry()
		Send.Append(paytobox, false)

		vb1 := ui.NewVerticalBox()
		vb1.SetPadded(true)
		vb1.Append(ui.NewLabel("Amount:"), false)
		amountbox := ui.NewSpinbox(0, 1<<30)
		vb1.Append(amountbox, false)

		vb2 := ui.NewVerticalBox()
		vb2.SetPadded(true)
		vb2.Append(ui.NewLabel("Fee:"), false)
		feebox := ui.NewSpinbox(0, 1<<30)
		vb2.Append(feebox, false)

		hb0 := ui.NewHorizontalBox()
		hb0.SetPadded(true)
		hb0.Append(vb1, true)
		hb0.Append(vb2, false)

		Send.Append(hb0, false)

		sendbut := ui.NewButton("Send")
		maxbut := ui.NewButton("Max")
		hb1 := ui.NewHorizontalBox()
		hb1.SetPadded(true)
		hb1.Append(sendbut, false)
		hb1.Append(maxbut, false)
		Send.Append(hb1, false)

		Earn := ui.NewVerticalBox()
		Earn.SetPadded(true)
		Earn.Append(ui.NewHorizontalSeparator(), false)
		Earn.Append(ui.NewLabel("Allow the client to create new blocks."), false)
		createnewblocksbox := ui.NewCheckbox("Create new blocks")
		usecpubox := ui.NewCheckbox("Use CPU")
		usecpubox.Disable()
		Earn.Append(createnewblocksbox, false)
		Earn.Append(usecpubox, false)

		tab := ui.NewTab()
		tab.Append("Receive", Space(Receive))
		tab.Append("Send", Space(Send))
		tab.Append("Earn coins", Space(Earn))

		//	tab.Append("Network",Network)

		vb := ui.NewVerticalBox()
		vb.SetPadded(true)
		vb.Append(tab, true)
		statusbar := ui.NewHorizontalBox()
		statusbar.SetPadded(true)
		balancestatus := ui.NewLabel("")
		networkstatus := ui.NewLabel("")
		statusbar.Append(ui.NewHorizontalSeparator(), false)
		statusbar.Append(balancestatus, true)
		statusbar.Append(networkstatus, false)
		statusbar.Append(ui.NewHorizontalSeparator(), false)

		vb.Append(statusbar, false)

		window := ui.NewWindow("Wallet", 500, 100, false)
		UpdateBalance := func() {
			go func() {
				mystate := core.GetBalance(string(addr))
				if mystate != nil {
					balance = mystate.Balance
					conf = mystate.LastBlockId
					ui.QueueMain(func() {
						if core.Main != nil {
							balancestatus.SetText(fmt.Sprintf("Balance: %d", balance) + subscriptnumber(core.Main.Higher.Id-conf))
						}
					})
				}
			}()
		}

		sendbut.OnClicked(func(self *ui.Button) {

			payto, err := hex.DecodeString(paytobox.Text())
			if err != nil || len(payto) != 32 {
				ui.MsgBox(window, "Error", "Invalid address")
				return
			}
			mystate := core.GetBalance(string(addr))

			tx := core.NewTx([][]byte{core.Pub}, 1)

			amount := uint64(amountbox.Value())
			fee := uint64(feebox.Value())

			if amount+fee > mystate.Balance {
				ui.MsgBox(window, "Error", "Insufficient funds")
				return
			}

			tx.AddOut(string(payto), amount)
			tx.Fee = fee
			tx.Nonce = mystate.Nonce
			tx.Sign(core.Priv)
			core.PushTx(tx)
			ui.MsgBox(window, "", "Transaction was sent")

		})

		maxbut.OnClicked(func(self *ui.Button) {
			amountbox.SetValue(int(balance))
		})

		createnewblocksbox.OnToggled(func(self *ui.Checkbox) {
			core.Mineblocks = self.Checked()
			if !self.Checked() {
				usecpubox.Disable()
				usecpubox.SetChecked(false)
				core.Minecpu = false
			} else {
				usecpubox.Enable()
			}
		})

		usecpubox.OnToggled(func(self *ui.Checkbox) {
			core.Minecpu = self.Checked()
		})
		hide.OnClicked(func(self *ui.Button) {
			if self.Text() == "Hide" {
				self.SetText("Show")
				privbox.SetText("")
				pubbox.SetText("")
				seedbox.SetText("")
			} else {
				self.SetText("Hide")
				privbox.SetText(fmt.Sprintf("%x", priv))
				pubbox.SetText(fmt.Sprintf("%x", pub))
				seedbox.SetText(string(seed))

			}
		})

		genacc.OnClicked(func(self *ui.Button) {
			if len(seedbox.Text()) < 20 {
				ui.MsgBox(window, "Error", "Length of seed should be more than 20")
				return
			}

			seed = []byte(seedbox.Text())

			addr, priv, pub = core.SetSeed(seed)
			addrbox.SetText(fmt.Sprintf("%x", addr))
			privbox.SetText(fmt.Sprintf("%x", priv))
			pubbox.SetText(fmt.Sprintf("%x", pub))

			UpdateBalance()
		})

		go func() {
			prevmain := core.Main
			for {
				if core.Main != nil && prevmain != core.Main {
					UpdateBalance()
					ui.QueueMain(func() {
						networkstatus.SetText(fmt.Sprintf("Network: %d:%6x", core.Main.Higher.Id, core.Main.Higher.Hash()[0:6]))
					})
				}
				if core.Main == nil {

					ui.QueueMain(func() {
						balancestatus.SetText("Balance: 0")
						networkstatus.SetText("Network: None")
					})
				}

				prevmain = core.Main
				time.Sleep(1 * time.Second)

			}
		}()

		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.SetChild(vb)
		window.Show()
	})
}

func subscriptnumber(i uint64) string {
	s := ""
	for ; i != 0; i = i / 10 {
		s = string([]byte{0xe2, 0x82, 0x80 + uint8(i%10)}) + s
	}
	return s
}

func Space(x ui.Control) ui.Control {
	fr0 := ui.NewHorizontalBox()
	fr0.SetPadded(true)
	fr0.Append(ui.NewHorizontalBox(), false)
	fr0.Append(x, true)
	fr0.Append(ui.NewHorizontalBox(), false)
	return fr0
}
