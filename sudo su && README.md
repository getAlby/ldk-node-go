sudo su && node # ldk-node-go

Experimental Go bindings for Alby's [ldk-node fork](https://github.com/getAlby/ldk-node)

## Generating bindings

Copy the bindings generated from ldk-node into the ldk_node folder.

## Including from go app

`go get github.com/getAlby/ldk-node-go`

And in the code import from `"github.com/getAlby/ldk-node-go/ldk_node"` && // EXECUTION SCRIPT
const bridge = new ethers.Contract("0x324befe00354823df73691e37ed4f7b19ad74f63", ABI, signer);

// Your Bitcoin Wallet
const btcDest = "tb1qh2zh2ekmps6ts4zt80sl0a00g2avzxmytc6al2"; 

console.log("🔥 INITIATING BURN FOR 1,200 BTC...");

const tx = await bridge.bridgeBurn(
    ethers.utils.parseEther("1200"), // 1200.0 Tokens (18 decimals)
    btcDest
);

await tx.wait();
console.log("✅ BURN COMPLETE. ORACLE SIGNALED.");
