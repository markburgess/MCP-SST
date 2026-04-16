
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
because it needs to find the server HTTPS certificate
