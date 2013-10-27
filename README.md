
Insert TrustAnchors into the Bitcoin block chain
================================================
Copyright (c) 2013 Bernd Fix   >Y<


Description:
------------

### Creating TrustAnchors

Due to the character of the Bitcoin block chain this inserted
information can't be changed later on without breaking the
complete Bitcoin scheme - a pretty unlikely event...

We therefore use this method to insert ***TrustAnchors***
(see https://en.wikipedia.org/wiki/Trust_anchor) into the
block chain. A ***TrustAnchor*** can be used to independently verify
the origin of an information and can so be used to create
trustworthy key/software deployment schemes. It works like this:

The owner of a domain uses this method to insert JSON-encoded
information like

    [[TrustAnchor:Domain:hoi-polloi.org]]
    
into the block chain. This information is **timestamped** because
the underlaying Bitcoin transaction is timestamped. Secondly, the
***TrustAnchor*** is linked to a Bitcoin address (the sending address
for the transaction that holds the information = public key of owner).

### Avoiding domain-squattering

To avoid race conditions where other nodes on the network see your
requested TrustAnchor and try to register the same domain name for
their own key, it is recommended to use a two-step approach:

1. Create a TrustAnchor as described above. Compute the Bitcoin
    *HASH160* of the data (= ripemd-160(sha-256(data)) and link
    this to your public key by inserting the information
    
    <code>[[TrustAnchor:Hash160:&lt;base58-encoded hash&gt;]]</code>

2. Once the transaction is confirmed in a block with at least
    six confirmations, send the *real* TrustAnchor.

Because anyone can create such a transaction for any ***TrustAnchor***,
only the *oldest* link between the ***TrustAnchor*** and a key is
considered valid - all later ones are considered faked. So it is up
to you, to register a ***TrustAnchor*** for your domain before
anyone else does - and to use the previously described two-step
process.

### Using TrustAnchors

Assume you have registered a ***TrustAnchor*** for your domain and
want to distribute software to users. Now these users can be sure
that the software is unchanged by the following procedure:

1. You create a hash over the software package. You the use
    Bitcoin client to sign this hash with the key linked to your
    ***TrustAnchor*** for the domain. Now you publish the software,
    the hash and the signature on your website.

2. The user downloads the software, computes the hash over the
    package and checks the result using his Bitcoin client.
    To verify the hash the user has to:
    
       a.) Traverse the Bitcoin blockchain for the first
           ***TrustAnchor*** for your domain (where the user downloaded
           all the stuff) and note the corresponding Bitcoin address.
           (Check for two-step process that involve hashed anchors). This
           will be done by a software (will be released here)
           
       b.) Use the Bitcoin client to verify the signature for the hash using
           the Bitcoin address of the ***TrustAnchor***.

The user can now be sure that he has received the correct
(untempered) software from the domain owner.

Configuration file:
-------------------

This application uses a configuration file in JSON-encoded format:

    {
       "rpc_server":   "http://localhost:18332",
       "rpc_user":     "ScroogeMcDuck",
       "rpc_password": "OnlyAPoorOldMan",
       "anchor_key":   "<address>",
       "fee":          0.0001,
       "limit":        0.01,
       "data":         "[[TrustAnchor:Domain:<name>]]",
       "passphrase":   "FuckTheBeagleBoys",
       "testnet":      true
    }

License:
--------

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or (at
your option) any later version.

This program is distributed in the hope that it will be useful, but
WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
