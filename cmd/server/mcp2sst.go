package main

import (
	"fmt"
	"os"
	"flag"
	"github.com/markburgess/MCP-SST/generated-server/mcptools" 
	"github.com/mark3labs/mcp-go/server"
)

// ***************************************************************

func main() {

	mcptools.SELF_SIGNED_CERT = Init()

	// 1. Create a new MCP server

	mcpserver := server.NewMCPServer("MCP-SSTorytime", "1.0.0")
	fmt.Println("Looking for a self-signed certificate at",mcptools.SELF_SIGNED_CERT)
	
	// 2. Register a tool (AddTool) with a handler function

	mcpserver.AddTool(mcptools.NewN4LqueryMCPTool(),mcptools.N4LqueryHandler)

	httpServer := server.NewStreamableHTTPServer(mcpserver)

	// Start the server

	fmt.Println("MCP server running on port 8888")
	
	if err := httpServer.Start(":8888"); err != nil {
		fmt.Println("Server failed to start: %v", err)
		os.Exit(-1)
	}
}


// ******************************************************************************

func Init() string {

	flag.Usage = Usage

	resourcePtr := flag.String("cert", "../server/cert.pem", "self-signed certicate path/name")

	flag.Parse()

	return *resourcePtr
}

//**************************************************************

func Usage() {

	fmt.Printf("usage: main [-cert path/to/https-cert]\n")
	os.Exit(1)
}

