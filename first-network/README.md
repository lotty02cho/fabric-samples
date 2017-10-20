## Build Your First Network (BYFN)

The directions for using this are documented in the Hyperledger Fabric
["Build Your First Network"](http://hyperledger-fabric.readthedocs.io/en/latest/build_network.html) tutorial.


<div role="main" class="document" itemscope="itemscope" itemtype="http://schema.org/Article">

<div itemprop="articleBody">

<div class="section" id="building-your-first-network">

# Building Your First Network[¶](#building-your-first-network "Permalink to this headline")

<div class="admonition note">

Note

These instructions have been verified to work against the version “1.0.0” tagged Docker images and the pre-compiled setup utilities within the supplied tar file. If you run these commands with images or tools from the current master branch, it is possible that you will see configuration and panic errors.

</div>

The build your first network (BYFN) scenario provisions a sample Hyperledger Fabric network consisting of two organizations, each maintaining two peer nodes, and a “solo” ordering service.

<div class="section" id="install-prerequisites">

## Install prerequisites[¶](#install-prerequisites "Permalink to this headline")

Before we begin, if you haven’t already done so, you may wish to check that you have all the [<span class="doc">Prerequisites</span>](prereqs.html) installed on the platform(s) on which you’ll be developing blockchain applications and/or operating Hyperledger Fabric.

You will also need to download and install the [<span class="doc">Hyperledger Fabric Samples</span>](samples.html). You will notice that there are a number of samples included in the `<span class="pre">fabric-samples</span>` repository. We will be using the `<span class="pre">first-network</span>` sample. Let’s open that sub-directory now.

<div class="code bash highlight-default">

<div class="highlight">

<pre><span></span><span class="n">cd</span> <span class="n">first</span><span class="o">-</span><span class="n">network</span>
</pre>

</div>

</div>

<div class="admonition note">

Note

The supplied commands in this documentation **MUST** be run from your `<span class="pre">first-network</span>` sub-directory of the `<span class="pre">fabric-samples</span>` repository clone. If you elect to run the commands from a different location, the various provided scripts will be unable to find the binaries.

</div>

</div>

<div class="section" id="want-to-run-it-now">

## Want to run it now?[¶](#want-to-run-it-now "Permalink to this headline")

We provide a fully annotated script - `<span class="pre">byfn.sh</span>` - that leverages these Docker images to quickly bootstrap a Hyperledger Fabric network comprised of 4 peers representing two different organizations, and an orderer node. It will also launch a container to run a scripted execution that will join peers to a channel, deploy and instantiate chaincode and drive execution of transactions against the deployed chaincode.

Here’s the help text for the `<span class="pre">byfn.sh</span>` script:

<div class="code bash highlight-default">

<div class="highlight">

<pre><span></span><span class="o">./</span><span class="n">byfn</span><span class="o">.</span><span class="n">sh</span> <span class="o">-</span><span class="n">h</span>
<span class="n">Usage</span><span class="p">:</span>
  <span class="n">byfn</span><span class="o">.</span><span class="n">sh</span> <span class="o">-</span><span class="n">m</span> <span class="n">up</span><span class="o">|</span><span class="n">down</span><span class="o">|</span><span class="n">restart</span><span class="o">|</span><span class="n">generate</span> <span class="p">[</span><span class="o">-</span><span class="n">c</span> <span class="o"><</span><span class="n">channel</span> <span class="n">name</span><span class="o">></span><span class="p">]</span> <span class="p">[</span><span class="o">-</span><span class="n">t</span> <span class="o"><</span><span class="n">timeout</span><span class="o">></span><span class="p">]</span>
  <span class="n">byfn</span><span class="o">.</span><span class="n">sh</span> <span class="o">-</span><span class="n">h</span><span class="o">|--</span><span class="n">help</span> <span class="p">(</span><span class="nb">print</span> <span class="n">this</span> <span class="n">message</span><span class="p">)</span>
    <span class="o">-</span><span class="n">m</span> <span class="o"><</span><span class="n">mode</span><span class="o">></span> <span class="o">-</span> <span class="n">one</span> <span class="n">of</span> <span class="s1">'up'</span><span class="p">,</span> <span class="s1">'down'</span><span class="p">,</span> <span class="s1">'restart'</span> <span class="ow">or</span> <span class="s1">'generate'</span>
      <span class="o">-</span> <span class="s1">'up'</span> <span class="o">-</span> <span class="n">bring</span> <span class="n">up</span> <span class="n">the</span> <span class="n">network</span> <span class="k">with</span> <span class="n">docker</span><span class="o">-</span><span class="n">compose</span> <span class="n">up</span>
      <span class="o">-</span> <span class="s1">'down'</span> <span class="o">-</span> <span class="n">clear</span> <span class="n">the</span> <span class="n">network</span> <span class="k">with</span> <span class="n">docker</span><span class="o">-</span><span class="n">compose</span> <span class="n">down</span>
      <span class="o">-</span> <span class="s1">'restart'</span> <span class="o">-</span> <span class="n">restart</span> <span class="n">the</span> <span class="n">network</span>
      <span class="o">-</span> <span class="s1">'generate'</span> <span class="o">-</span> <span class="n">generate</span> <span class="n">required</span> <span class="n">certificates</span> <span class="ow">and</span> <span class="n">genesis</span> <span class="n">block</span>
    <span class="o">-</span><span class="n">c</span> <span class="o"><</span><span class="n">channel</span> <span class="n">name</span><span class="o">></span> <span class="o">-</span> <span class="n">config</span> <span class="n">name</span> <span class="n">to</span> <span class="n">use</span> <span class="p">(</span><span class="n">defaults</span> <span class="n">to</span> <span class="s2">"mychannel"</span><span class="p">)</span>
    <span class="o">-</span><span class="n">t</span> <span class="o"><</span><span class="n">timeout</span><span class="o">></span> <span class="o">-</span> <span class="n">CLI</span> <span class="n">timeout</span> <span class="n">duration</span> <span class="ow">in</span> <span class="n">microseconds</span> <span class="p">(</span><span class="n">defaults</span> <span class="n">to</span> <span class="mi">10000</span><span class="p">)</span>

<span class="n">Typically</span><span class="p">,</span> <span class="n">one</span> <span class="n">would</span> <span class="n">first</span> <span class="n">generate</span> <span class="n">the</span> <span class="n">required</span> <span class="n">certificates</span> <span class="ow">and</span>
<span class="n">genesis</span> <span class="n">block</span><span class="p">,</span> <span class="n">then</span> <span class="n">bring</span> <span class="n">up</span> <span class="n">the</span> <span class="n">network</span><span class="o">.</span> <span class="n">e</span><span class="o">.</span><span class="n">g</span><span class="o">.</span><span class="p">:</span>

  <span class="n">byfn</span><span class="o">.</span><span class="n">sh</span> <span class="o">-</span><span class="n">m</span> <span class="n">generate</span> <span class="o">-</span><span class="n">c</span> <span class="o"><</span><span class="n">channelname</span><span class="o">></span>
  <span class="n">byfn</span><span class="o">.</span><span class="n">sh</span> <span class="o">-</span><span class="n">m</span> <span class="n">up</span> <span class="o">-</span><span class="n">c</span> <span class="o"><</span><span class="n">channelname</span><span class="o">></span>
</pre>

</div>

</div>

If you choose not to supply a channel name, then the script will use a default name of `<span class="pre">mychannel</span>`. The CLI timeout parameter (specified with the -t flag) is an optional value; if you choose not to set it, then your CLI container will exit upon conclusion of the script.

<div class="section" id="generate-network-artifacts">

### Generate Network Artifacts[¶](#generate-network-artifacts "Permalink to this headline")

Ready to give it a go? Okay then! Execute the following command:

<div class="code bash highlight-default">

<div class="highlight">

<pre><span></span><span class="o">./</span><span class="n">byfn</span><span class="o">.</span><span class="n">sh</span> <span class="o">-</span><span class="n">m</span> <span class="n">generate</span>
</pre>

</div>

</div>

You will see a brief description as to what will occur, along with a yes/no command line prompt. Respond with a `<span class="pre">y</span>` to execute the described action.

<div class="code bash highlight-default">

<div class="highlight">

<pre><span></span>Generating certs and genesis block for with channel 'mychannel' and CLI timeout of '10000'
Continue (y/n)?y
proceeding ...
/Users/xxx/dev/fabric-samples/bin/cryptogen

##########################################################
##### Generate certificates using cryptogen tool #########
##########################################################
org1.example.com
2017-06-12 21:01:37.334 EDT [bccsp] GetDefault -> WARN 001 Before using BCCSP, please call InitFactories(). Falling back to bootBCCSP.
...

/Users/xxx/dev/fabric-samples/bin/configtxgen
##########################################################
#########  Generating Orderer Genesis block ##############
##########################################################
2017-06-12 21:01:37.558 EDT [common/configtx/tool] main -> INFO 001 Loading configuration
2017-06-12 21:01:37.562 EDT [msp] getMspConfig -> INFO 002 intermediate certs folder not found at [/Users/xxx/dev/byfn/crypto-config/ordererOrganizations/example.com/msp/intermediatecerts]. Skipping.: [stat /Users/xxx/dev/byfn/crypto-config/ordererOrganizations/example.com/msp/intermediatecerts: no such file or directory]
...
2017-06-12 21:01:37.588 EDT [common/configtx/tool] doOutputBlock -> INFO 00b Generating genesis block
2017-06-12 21:01:37.590 EDT [common/configtx/tool] doOutputBlock -> INFO 00c Writing genesis block

#################################################################
### Generating channel configuration transaction 'channel.tx' ###
#################################################################
2017-06-12 21:01:37.634 EDT [common/configtx/tool] main -> INFO 001 Loading configuration
2017-06-12 21:01:37.644 EDT [common/configtx/tool] doOutputChannelCreateTx -> INFO 002 Generating new channel configtx
2017-06-12 21:01:37.645 EDT [common/configtx/tool] doOutputChannelCreateTx -> INFO 003 Writing new channel tx

#################################################################
#######    Generating anchor peer update for Org1MSP   ##########
#################################################################
2017-06-12 21:01:37.674 EDT [common/configtx/tool] main -> INFO 001 Loading configuration
2017-06-12 21:01:37.678 EDT [common/configtx/tool] doOutputAnchorPeersUpdate -> INFO 002 Generating anchor peer update
2017-06-12 21:01:37.679 EDT [common/configtx/tool] doOutputAnchorPeersUpdate -> INFO 003 Writing anchor peer update

#################################################################
#######    Generating anchor peer update for Org2MSP   ##########
#################################################################
2017-06-12 21:01:37.700 EDT [common/configtx/tool] main -> INFO 001 Loading configuration
2017-06-12 21:01:37.704 EDT [common/configtx/tool] doOutputAnchorPeersUpdate -> INFO 002 Generating anchor peer update
2017-06-12 21:01:37.704 EDT [common/configtx/tool] doOutputAnchorPeersUpdate -> INFO 003 Writing anchor peer update
</pre>

</div>

</div>

This first step generates all of the certificates and keys for all our various network entities, the `<span class="pre">genesis</span> <span class="pre">block</span>` used to bootstrap the ordering service, and a collection of configuration transactions required to configure a [<span class="std std-ref">Channel</span>](glossary.html#channel).

</div>

<div class="section" id="bring-up-the-network">

### Bring Up the Network[¶](#bring-up-the-network "Permalink to this headline")

Next, you can bring the network up with the following command:

<div class="code bash highlight-default">

<div class="highlight">

<pre><span></span><span class="o">./</span><span class="n">byfn</span><span class="o">.</span><span class="n">sh</span> <span class="o">-</span><span class="n">m</span> <span class="n">up</span>
</pre>

</div>

</div>

Once again, you will be prompted as to whether you wish to continue or abort. Respond with a `<span class="pre">y</span>`:

<div class="code bash highlight-default">

<div class="highlight">

<pre><span></span>Starting with channel 'mychannel' and CLI timeout of '10000'
Continue (y/n)?y
proceeding ...
Creating network "net_byfn" with the default driver
Creating peer0.org1.example.com
Creating peer1.org1.example.com
Creating peer0.org2.example.com
Creating orderer.example.com
Creating peer1.org2.example.com
Creating cli

 ____    _____      _      ____    _____
/ ___|  |_   _|    / \    |  _ \  |_   _|
\___ \    | |     / _ \   | |_) |   | |
 ___) |   | |    / ___ \  |  _ <    | |
|____/    |_|   /_/   \_\ |_| \_\   |_|

Channel name : mychannel
Creating channel...
</pre>

</div>

</div>

The logs will continue from there. This will launch all of the containers, and then drive a complete end-to-end application scenario. Upon successful completion, it should report the following in your terminal window:

<div class="code bash highlight-default">

<div class="highlight">

<pre><span></span><span class="mi">2017</span><span class="o">-</span><span class="mi">05</span><span class="o">-</span><span class="mi">16</span> <span class="mi">17</span><span class="p">:</span><span class="mi">08</span><span class="p">:</span><span class="mf">01.366</span> <span class="n">UTC</span> <span class="p">[</span><span class="n">msp</span><span class="p">]</span> <span class="n">GetLocalMSP</span> <span class="o">-></span> <span class="n">DEBU</span> <span class="mi">004</span> <span class="n">Returning</span> <span class="n">existing</span> <span class="n">local</span> <span class="n">MSP</span>
<span class="mi">2017</span><span class="o">-</span><span class="mi">05</span><span class="o">-</span><span class="mi">16</span> <span class="mi">17</span><span class="p">:</span><span class="mi">08</span><span class="p">:</span><span class="mf">01.366</span> <span class="n">UTC</span> <span class="p">[</span><span class="n">msp</span><span class="p">]</span> <span class="n">GetDefaultSigningIdentity</span> <span class="o">-></span> <span class="n">DEBU</span> <span class="mi">005</span> <span class="n">Obtaining</span> <span class="n">default</span> <span class="n">signing</span> <span class="n">identity</span>
<span class="mi">2017</span><span class="o">-</span><span class="mi">05</span><span class="o">-</span><span class="mi">16</span> <span class="mi">17</span><span class="p">:</span><span class="mi">08</span><span class="p">:</span><span class="mf">01.366</span> <span class="n">UTC</span> <span class="p">[</span><span class="n">msp</span><span class="o">/</span><span class="n">identity</span><span class="p">]</span> <span class="n">Sign</span> <span class="o">-></span> <span class="n">DEBU</span> <span class="mi">006</span> <span class="n">Sign</span><span class="p">:</span> <span class="n">plaintext</span><span class="p">:</span> <span class="mi">0</span><span class="n">AB1070A6708031A0C08F1E3ECC80510</span><span class="o">...</span><span class="mi">6</span><span class="n">D7963631A0A0A0571756572790A0161</span>
<span class="mi">2017</span><span class="o">-</span><span class="mi">05</span><span class="o">-</span><span class="mi">16</span> <span class="mi">17</span><span class="p">:</span><span class="mi">08</span><span class="p">:</span><span class="mf">01.367</span> <span class="n">UTC</span> <span class="p">[</span><span class="n">msp</span><span class="o">/</span><span class="n">identity</span><span class="p">]</span> <span class="n">Sign</span> <span class="o">-></span> <span class="n">DEBU</span> <span class="mi">007</span> <span class="n">Sign</span><span class="p">:</span> <span class="n">digest</span><span class="p">:</span> <span class="n">E61DB37F4E8B0D32C9FE10E3936BA9B8CD278FAA1F3320B08712164248285C54</span>
<span class="n">Query</span> <span class="n">Result</span><span class="p">:</span> <span class="mi">90</span>
<span class="mi">2017</span><span class="o">-</span><span class="mi">05</span><span class="o">-</span><span class="mi">16</span> <span class="mi">17</span><span class="p">:</span><span class="mi">08</span><span class="p">:</span><span class="mf">15.158</span> <span class="n">UTC</span> <span class="p">[</span><span class="n">main</span><span class="p">]</span> <span class="n">main</span> <span class="o">-></span> <span class="n">INFO</span> <span class="mi">008</span> <span class="n">Exiting</span><span class="o">.....</span>
<span class="o">=====================</span> <span class="n">Query</span> <span class="n">on</span> <span class="n">PEER3</span> <span class="n">on</span> <span class="n">channel</span> <span class="s1">'mychannel'</span> <span class="ow">is</span> <span class="n">successful</span> <span class="o">=====================</span>

<span class="o">=====================</span> <span class="n">All</span> <span class="n">GOOD</span><span class="p">,</span> <span class="n">BYFN</span> <span class="n">execution</span> <span class="n">completed</span> <span class="o">=====================</span>

 <span class="n">_____</span>   <span class="n">_</span>   <span class="n">_</span>   <span class="n">____</span>
<span class="o">|</span> <span class="n">____</span><span class="o">|</span> <span class="o">|</span> \ <span class="o">|</span> <span class="o">|</span> <span class="o">|</span>  <span class="n">_</span> \
<span class="o">|</span>  <span class="n">_</span><span class="o">|</span>   <span class="o">|</span>  \<span class="o">|</span> <span class="o">|</span> <span class="o">|</span> <span class="o">|</span> <span class="o">|</span> <span class="o">|</span>
<span class="o">|</span> <span class="o">|</span><span class="n">___</span>  <span class="o">|</span> <span class="o">|</span>\  <span class="o">|</span> <span class="o">|</span> <span class="o">|</span><span class="n">_</span><span class="o">|</span> <span class="o">|</span>
<span class="o">|</span><span class="n">_____</span><span class="o">|</span> <span class="o">|</span><span class="n">_</span><span class="o">|</span> \<span class="n">_</span><span class="o">|</span> <span class="o">|</span><span class="n">____</span><span class="o">/</span>
</pre>

</div>

</div>

You can scroll through these logs to see the various transactions. If you don’t get this result, then jump down to the [<span class="std std-ref">Troubleshooting</span>](#troubleshoot) section and let’s see whether we can help you discover what went wrong.

</div>

<div class="section" id="bring-down-the-network">

### Bring Down the Network[¶](#bring-down-the-network "Permalink to this headline")

Finally, let’s bring it all down so we can explore the network setup one step at a time. The following will kill your containers, remove the crypto material and four artifacts, and delete the chaincode images from your Docker Registry:

<div class="code bash highlight-default">

<div class="highlight">

<pre><span></span><span class="o">./</span><span class="n">byfn</span><span class="o">.</span><span class="n">sh</span> <span class="o">-</span><span class="n">m</span> <span class="n">down</span>
</pre>

</div>

</div>

Once again, you will be prompted to continue, respond with a `<span class="pre">y</span>`:

<div class="code bash highlight-default">

<div class="highlight">

<pre><span></span>Stopping with channel 'mychannel' and CLI timeout of '10000'
Continue (y/n)?y
proceeding ...
WARNING: The CHANNEL_NAME variable is not set. Defaulting to a blank string.
WARNING: The TIMEOUT variable is not set. Defaulting to a blank string.
Removing network net_byfn
468aaa6201ed
...
Untagged: dev-peer1.org2.example.com-mycc-1.0:latest
Deleted: sha256:ed3230614e64e1c83e510c0c282e982d2b06d148b1c498bbdcc429e2b2531e91
...
</pre>

</div>

</div>

If you’d like to learn more about the underlying tooling and bootstrap mechanics, continue reading. In these next sections we’ll walk through the various steps and requirements to build a fully-functional Hyperledger Fabric network.

</div>

</div>

<div class="section" id="crypto-generator">
