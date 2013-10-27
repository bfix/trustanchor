/*
 * Insert extra information into the Bitcoin block chain.
 * ======================================================
 *
 * Due to the character of the Bitcoin block chain this inserted
 * information can't be changed later on without breaking the
 * complete Bitcoin scheme - a pretty unlikely event...
 *
 * We therefore use this method to insert "TrustAnchors" into the
 * block chain. A "TrustAnchor" can be used to independently verify
 * the origin of an information and can so be used to create
 * trustworthy key/software deployment schemes. That works like thos:
 *
 * The owner of a domain can use this method for example to insert
 * JSON-encoded information like
 *
 *        [[TrustAnchor:Domain:hoi-polloi.org]]
 *
 * into the block chain. This information is "timestamped" because
 * the underlaying Bitcoin transaction is timestamped. Secondly, the
 * "TrustAnchor" is linked to a Bitcoin address (the receiving address
 " for the transaction that holds the information).
 *
 * To avoid race conditions where other nodes on the network see you
 * requested TrustAnchor and try to register the same domain name for
 * their own key, it is recommended to use a two-step approach:
 *
 * (1) Create a TrustAnchor as described above. Compute the Bitcoin
 *     HASH160 of the data (= ripemd-160(sha-256(data)) and link
 *     this to your public key by inserting the information
 *
 *      [[TrustAnchor:Hash160:<base58-encoded hash>]]
 *
 * (2) Once the transaction is confirmed in a block with at least
 *     six confirmations, send the "real" TrustAnchor.
 *
 * Because anyone can create such a transaction for any "TrustAnchor",
 * only the *oldest* link between the "TrustAnchor" and a key is
 * considered valid - all later ones are considered faked. So it is up
 * to you, to register a "TrustAnchor" for your information before
 * anyone else does - and to use the previously described two-step
 * process.
 *
 * Assume you have registered a "TrustAnchor" for your domain and
 * want to distribute software to users. Now these users can be sure
 * that the software is unchanged by the following procedure:
 *
 * (1) You create a hash over the spftware package. You the use
 *     Bitcoin to sign this hash with the key linked to your
 *     "TrustAnchor" for the domain. Now you publish the software,
 *     the hash and the signature on your website.
 *
 * (2) The user downloads the software, computes the hash over the
 *     package and checks the result. To verify the hash the user
 *     has to:
 *        (a) Traverse the Bitcoin blockchain for the first
 *            "TrustAnchor" for your domain (where the user downloaded
 *            all the stuff) and note the corresponding Bitcoin address.
 *            (Check for two-step process that involve hashed anchors)
 *        (c) Use Bitcoin to verify the signature for the hash using
 *            the Bitcoin address of the "TrustAnchor".
 *
 * The user can now be sure that he has received the correct
 * (untempered) software from the domain owner.
 *
 * This application uses a configuration file in JSON-encoded format:
 *
 * {
 *    "rpc_server":   "http://localhost:18332",
 *    "rpc_user":     "ScroogeMcDuck",
 *    "rpc_password": "OnlyAPoorOldMan",
 *    "anchor_key":   "<address>",
 *    "fee":          0.0001,
 *    "limit":        0.01,
 *    "data":         "[[TrustAnchor:Domain:<name>]]",
 *    "passphrase":   "FuckTheBeagleBoys",
 *    "testnet":      true
 * }
 *
 * (c) 2013 Bernd Fix   >Y<
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

///////////////////////////////////////////////////////////////////////
// import external declaratiuons

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bfix/gospel/bitcoin/rpc"
	"github.com/bfix/gospel/bitcoin/util"
	"os"
)

///////////////////////////////////////////////////////////////////////
// constants

const (
	VERBOSE = true
)

///////////////////////////////////////////////////////////////////////
// injection configuration data

type Config struct {
	Server     string  `json:"rpc_server"`
	User       string  `json:"rpc_user"`
	Password   string  `json:"rpc_password"`
	AnchorKey  string  `json:"anchor_key"`
	Receiver   string  `jsdon:"receiver"`
	Fee        float64 `json:"fee"`
	Limit      float64 `json:"limit"`
	Data       string  `json:"data"`
	Passphrase string  `json:"passphrase"`
	TestNet    bool    `json:"testnet"`
}

///////////////////////////////////////////////////////////////////////
// application entry point

func main() {
	flag.Parse()

	//-----------------------------------------------------------------
	// access config data
	fName := "inject.json"
	switch {
	case flag.NArg() > 1:
		fmt.Println("Ignoring extraneous arguments")
		fallthrough
	case flag.NArg() == 1:
		fName = flag.Arg(0)
	default:
		fmt.Println("Using default injection file '" + fName + "'")
	}
	f, err := os.Open(fName)
	if err != nil {
		fmt.Println("Can't open file '" + fName + "': " + err.Error())
		return
	}
	defer f.Close()

	// decode config data
	dec := json.NewDecoder(f)
	config := new(Config)
	if err = dec.Decode(config); err != nil {
		fmt.Println("Error decoding config file: " + err.Error())
		return
	}

	// read passphrase from console if not defined in config file
	if len(config.Passphrase) == 0 {
		fmt.Print("Enter passphrase for wallet: ")
		rdr := bufio.NewReader(os.Stdin)
		data, _, err := rdr.ReadLine()
		if err != nil {
			fmt.Println("Error reading input: " + err.Error())
			return
		}
		config.Passphrase = string(data)
	}

	//-----------------------------------------------------------------

	sess, err := rpc.NewSession(config.Server, config.User, config.Password)
	if err != nil {
		fmt.Println("session creation failed: " + err.Error())
		return
	}

	// unlock wallet
	err = sess.WalletPassphrase(config.Passphrase, 600)
	if err != nil {
		fmt.Println("WalletPassphrase(): " + err.Error())
	}
	defer func() {
		// lock wallet
		err = sess.WalletLock()
		if err != nil {
			fmt.Println("WalletLock(): " + err.Error())
		}
	}()

	//-----------------------------------------------------------------
	// check if we can sent from sender address at all
	rcv, err := sess.ListReceivedByAddress(6, false)
	if err != nil {
		fmt.Println("ListReceivedByAddress() failed: " + err.Error())
		return
	}
	fail := true
	for _, r := range rcv {
		if VERBOSE {
			fmt.Println("ListReceivedByAddress(): " + r.Address)
		}
		if r.Address == config.AnchorKey {
			fail = false
			break
		}
	}
	if fail {
		fmt.Println("can't send from address '" + config.AnchorKey + "'")
		return
	}

	//-----------------------------------------------------------------
	// traverse list of unspent transactions and find the smalled amount
	// that is equal or larger than the amount we want to spend (incl. fees).
	// make sure, unspend transactions belong to sender address
	unspent, err := sess.ListUnspent(6, 999999)
	if err != nil {
		fmt.Println("ListUnspent() failed: " + err.Error())
		return
	}
	if len(unspent) == 0 {
		fmt.Println("no bitcoins to spent in this wallet!")
		return
	}
	var (
		txid       string
		vout       int = -1
		balance    float64
		lastAmount float64 = 999999
	)
	for _, u := range unspent {
		// extract address
		script, err := hex.DecodeString(u.ScriptPubKey)
		if err != nil {
			fmt.Println("failed to decode script hex")
			return
		}
		if script[0] == 0x76 && script[1] == 0xa9 && script[2] == 0x14 {
			// recalc address
			if config.TestNet {
				script[2] = 111
			} else {
				script[2] = 0
			}
			hash := script[2:23]
			cs := util.Hash256(hash)
			hash = append(hash, cs[:4]...)
			addr := util.Base58Encode(hash)

			// check if sender address
			if VERBOSE {
				fmt.Println("ListUnspent(): " + addr)
			}
			if addr != config.AnchorKey {
				continue
			}
		} else {
			continue
		}

		if u.Amount > 2*config.Fee && (vout < 0 || u.Amount < lastAmount) {
			lastAmount = u.Amount
			vout = u.Output.Vout
			txid = u.Output.Id
			balance = u.Amount
		}
	}
	if vout < 0 {
		fmt.Println("not enough bitcoins to spent on transaction!")
		return
	}
	if balance > config.Limit {
		fmt.Println("smallest balance exceed limit!")
		return
	}

	//=================================================================
	// Transfer funds to provable prunable output
	//=================================================================

	outs := []rpc.Output{
		rpc.Output{
			Id:   txid,
			Vout: vout,
		},
	}
	ins := []rpc.Balance{
		rpc.Balance{
			Address: config.AnchorKey,
			Amount:  balance - config.Fee,
		},
	}
	raw, err := sess.CreateRawTransaction(outs, ins)
	if err != nil {
		fmt.Println("CreateRawTransaction() failed: " + err.Error())
		return
	}

	script, err := util.NullDataScript([]byte(config.Data))
	if err != nil {
		fmt.Println("NullDataScript() failed: " + err.Error())
		return
	}
	raw, err = util.ReplaceScriptPubKey(raw, script)
	if err != nil {
		fmt.Println("ReplaceScriptPubKey() failed: " + err.Error())
		return
	}

	// fill scriptSig to link to unspent transaction slot
	raw, complete, err := sess.SignRawTransaction(raw, nil, nil, "ALL")
	if err != nil {
		fmt.Println("SignRawTransaction(): " + err.Error())
		return
	}
	if !complete {
		fmt.Println("SignRawTransaction() not complete")
	}

	if VERBOSE {
		fmt.Println("===============================================================")
		fmt.Println(raw)
		fmt.Println("===============================================================")
		obj, _ := sess.DecodeRawTransactionAsObject(raw)
		fmt.Printf("%v\n", obj)
		fmt.Println("===============================================================")
	}

	// send transaction
	err = sess.SendRawTransaction(raw)
	if err != nil {
		fmt.Println("SendRawTransaction(): " + err.Error())
		return
	}
}
