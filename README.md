
# Trying to build an MCP to SST proxy

There are conflicting versions, muddled together by LLM online, but this finally compiles

<pre>
git clone https://github.com/markburgess/MCP-SST
cd MCP-SST
make
</pre>

This should build

The server needs to be run from the same bin directory as the SSTorytime
<pre>
  bin/http_server
</pre>
else specify the self-signed certificate file with
<pre>
bin/main -cert path/to/cert.pem

</pre>
