package mcptools

import (
	"net/http"
	"net/url"
	"io"
	"context"
	"fmt"
	"crypto/tls"
	"crypto/x509"
	"os"
	"encoding/json"
	"github.com/mark3labs/mcp-go/mcp"
)

var SELF_SIGNED_CERT string

// ******************************************************************************
// Input Schema for the N4Lquery tool - this is where we have to add tooltips
// ******************************************************************************

const help_hint_text = `Search the N4L knowledge graph with command CLI. Reserved commands start with \ backslash
Example: * Search for a precise word

The strings you type are normally treated as potential substrings to match within words.
If you want to insist a precise word match with nothing else included, i.e. the search term is
the entire node string,then you can usethe pling/bang/exclamation character !on both sides of the term,
or the vertical bar "pipe" symbol (which is not confused with the ! not operator):
<pre>
   !a1!
   |a1|
   "|deep purple|"              (exact match with space needs quotes!)
   "ephemeral or persistent"
</pre>

* Search with spaces in the string

If your search term contains spaces, exclose them in double quotes or use the <-> search operator (belonging to postgres *ts_vector*). If want to use logical operators to select or exclude certain words (or find matches based in related/derivative words) then the algorithm uses the ts_vector mathods and searching is by exact words. Then you need to use the substitute space <-> and <N> (not integer N) to represent spaces
<pre>
  strange<->kind<->of<->woman  // neighbouring lexemes (separated by space)
  strange<2>woman              // skip 2 lexemes
</pre>
(NB: the ts_vector method ignores insignificant words like "a", "in", "of", etc, so it will tend to ignore these
if you include them in a search string.)

If you simply want a (sub)string match, character by character, then quote the string:
<pre>
  "fish soup"
</pre>
This treats "fish soup" as a single possible substring, rather than as "fish" OR "soup".

* Search for any OR combination of a set of words

If your words are implicitly ORed together, then just separate by spaces.
<pre>
  word1 word2 word3
  recipe fish soup
</pre>
Conversely, words separated by spaces are ORed together.

* Search with logical expressions

You can use "& = AND", "! = NOT", "| = OR" in expressions, i.e. postgres ts_vector search logic in search terms, if you place them in quotes. This is very powerful. Notice that the !character is also used for hard-delimiting of strings. You might need to enclose your expression in quotes to keep it together.
<pre>
 a1&!b6
 "a1 & !b6"
 brain&!notes
 pink<->flo:*    // the :* operator completes a word starting with the prefix
</pre>
Note that, without the quotes, the latter string would  be understood as "a1 OR & OR NOT b6".

Postgres ts_vector() functionality is very powerful, but it relies on a dictionary. Currently only English language is supported. Hopefully this will change in the future.

* Search for any combination of a set of words in a chapter
<pre>
  word1 word2 word3 \chapter dictionary
  recipe fish soup  \chapter "my recipes"
</pre>

* Search for any combination of a set of words in named context, any chapter
<pre>
  word1 word2 word3 \context "weird words"
  recipe fish soup  \chapter food
</pre>

* General word search

<pre>
  word1 word2 word3 \chapter "my chapter" yourchapter \context "weird words"
  recipe fish soup  \chapter food \context food, recipes, dishes
</pre>

## Table of contents
<pre>
\toc
\map
\chapters
\chapter mychapter
</pre>

## Notes

* Print original notes from a chapter
<pre>
\notes mychapter
</pre>

* Print original notes from a context
<pre>
\notes \context mycontext
</pre>

* Print original notes not seen in the last four hours
<pre>
\notes chapter \new
</pre>

* Print original notes never seen
<pre>
\notes chapter \never
</pre>

## Stories and sequences
<pre>
\story (wuya)  # for unaccenting unicode characters
\story Mary
\sequence "create a pod"
\seq any \in \chapter kubernetes
\story any \chapter moon
</pre>

## Path solutions

<pre>
\paths \from start \to target
\from !a1! \to b6
</pre>

## Look for an arrow

<pre>
\arrow succeed
\arrow 1
\arrow 232
</pre>


## Look for concepts or terms extracted from text with text2N4L

Three synonyms are provided for convenience:

<pre>
\dna \chapter NDA                   # Show all terms discovered
\terms data patent \chapter nda     # show terms matching the list
\terms \chapter Darwin
\concepts \chapter moby
</pre>`


// ******************************************************************************


const N4LqueryInputSchema = `{
  "properties": {
    "body": {
      "properties": {
        "chapcontext": {
          "description": "string used for web group updates",
          "type": "string"
        },
        "name": {

            "description": "` + help_hint_text + `",

          "type": "string"
        },
        "nclass": {
          "description": "NPtr string classification for direct lookup",
          "type": "string"
        },
        "ncptr": {
          "description": "NPtr hash number for direct lookup",
          "type": "string"
        }
      },
      "required": [
        "name"
      ],
      "type": "object"
    }
  },
  "required": [
    "body"
  ],
  "type": "object"
}`

// Response Template for the N4Lquery tool (Status: 200, Content-Type: application/json)

const N4LqueryResponseTemplate_A = `# API Response Information

Below is the response template for this API endpoint.

The template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.

**Status Code:** 200

**Content-Type:** application/json

> Default for all non error reponses

## Response Structure

- Structure (Type: object):
  - **Response** (Type: string):
  - **Content** (Type: Combinator):
    - **One Of the following structures**:
      - **Option 1**: An array of nearest neighbours of the selected knowable text (Type: array):
        - **Items** (Type: array):
          - **Items**: An element in orbit around the knowable text (Type: object):
            - **Ctx**: The context key words of the orbiting item (Type: string):
            - **Dst**: An internal direct reference to a knowable item (Type: object):
              - **CPtr** (Type: integer):
              - **Class** (Type: integer):
            - **Text**: The dereferenced text of the orbiting item (Type: string):
            - **Wgt**: The link weight of the arrow (Type: number):
            - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
              - **Lat** (Type: number):
              - **Lon** (Type: number):
              - **R** (Type: number):
              - **X** (Type: number):
              - **Y** (Type: number):
              - **Z** (Type: number):
            - **STindex**: The Semantic Spacetime type (0-6) of the arrow (Type: integer):
            - **Arrow**: The arrow type relating the orbiting item to the base knowable text (Type: string):
            - **OOO**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
              - **R** (Type: number):
              - **X** (Type: number):
              - **Y** (Type: number):
              - **Z** (Type: number):
              - **Lat** (Type: number):
              - **Lon** (Type: number):
            - **Radius**: The distance of the related item from the base knowable text (Type: integer):
      - **Option 2**: A structure listing paths from a root node (Type: object):
        - **SuperNodes**: A list of nodes identified by their symmetrical roles in the paths (Type: array):
          - **Items** (Type: string):
        - **Title**: The text of the root node (Type: string):
        - **BTWC**: An array of Betweenness Centrality scores for the path matrix (Type: array):
          - **Items** (Type: string):
        - **Paths**: An array of paths starting from the root node (Type: array):
          - **Items** (Type: array):
            - **Items**: A curated journey through the knowable graph (Type: object):
              - **NPtr**: An internal direct reference to a knowable item (Type: object):
                - **CPtr** (Type: integer):
                - **Class** (Type: integer):
              - **STindex**: The Semantic Spacetime (0-6) type of the arrow (Type: integer):
              - **Chp**: The chapter the text belongs to (Type: string):
              - **Wgt**: The weight assigned to the arrow link (Type: number):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **R** (Type: number):
                - **X** (Type: number):
                - **Y** (Type: number):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
              - **Arr**: The arrow type of the next leg of the path (Type: integer):
              - **Ctx**: The context keywords the text belongs to (Type: string):
              - **Line**: The line number reference of the text in the original N4L notes (Type: integer):
              - **Name**: The text for this location in the path (Type: string):
        - **RootNode**: An internal direct reference to a knowable item (Type: object):
          - **CPtr** (Type: integer):
          - **Class** (Type: integer):
      - **Option 3**: A structure listing paths from a root node (Type: object):
        - **SuperNodes**: A list of nodes identified by their symmetrical roles in the paths (Type: array):
          - **Items** (Type: string):
        - **Title**: The text of the root node (Type: string):
        - **BTWC**: An array of Betweenness Centrality scores for the path matrix (Type: array):
          - **Items** (Type: string):
        - **Paths**: An array of paths starting from the root node (Type: array):
          - **Items** (Type: array):
            - **Items**: A curated journey through the knowable graph (Type: object):
              - **Wgt**: The weight assigned to the arrow link (Type: number):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **Y** (Type: number):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
                - **R** (Type: number):
                - **X** (Type: number):
              - **Arr**: The arrow type of the next leg of the path (Type: integer):
              - **Ctx**: The context keywords the text belongs to (Type: string):
              - **Line**: The line number reference of the text in the original N4L notes (Type: integer):
              - **Name**: The text for this location in the path (Type: string):
              - **NPtr**: An internal direct reference to a knowable item (Type: object):
                - **Class** (Type: integer):
                - **CPtr** (Type: integer):
              - **STindex**: The Semantic Spacetime (0-6) type of the arrow (Type: integer):
              - **Chp**: The chapter the text belongs to (Type: string):
        - **RootNode**: An internal direct reference to a knowable item (Type: object):
          - **CPtr** (Type: integer):
          - **Class** (Type: integer):
      - **Option 4**: An array of NodeEvents forming a story sequence about the search topic (Type: array):
        - **Items**: The content pointed to by a NodePtr (Type: object):
          - **L**: The string length of the text (Type: integer):
          - **NPtr**: An internal direct reference to a knowable item (Type: object):
            - **CPtr** (Type: integer):
            - **Class** (Type: integer):
          - **Text**: The knowable data (Type: string):
          - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
            - **X** (Type: number):
            - **Y** (Type: number):
            - **Z** (Type: number):
            - **Lat** (Type: number):
            - **Lon** (Type: number):
            - **R** (Type: number):
          - **Chap**: The chapter containing the text (Type: string):
          - **Context**: The context keywords explaining the relevance of the text (Type: string):
      - **Option 5**: A literal rendering of the original N4L page notes (Type: object):
        - **Context**: The context keywords labelling the knowable item (Type: string):
        - **Notes**: An array of the related notes matching the knowable item (Type: array):
          - **Items** (Type: array):
            - **Items**: A curated journey through the knowable graph (Type: object):
              - **Line**: The line number reference of the text in the original N4L notes (Type: integer):
              - **Name**: The text for this location in the path (Type: string):
              - **NPtr**: An internal direct reference to a knowable item (Type: object):
                - **CPtr** (Type: integer):
                - **Class** (Type: integer):
              - **STindex**: The Semantic Spacetime (0-6) type of the arrow (Type: integer):
              - **Chp**: The chapter the text belongs to (Type: string):
              - **Wgt**: The weight assigned to the arrow link (Type: number):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **R** (Type: number):
                - **X** (Type: number):
                - **Y** (Type: number):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
              - **Arr**: The arrow type of the next leg of the path (Type: integer):
              - **Ctx**: The context keywords the text belongs to (Type: string):
        - **Title**: The knowable item (Type: string):
      - **Option 6**: A table of contents matching chapters and contexts (Type: array):
        - **Items** (Type: object):
          - **Context** (Type: array):
            - **Items** (Type: object):
              - **Reln** (Type: array):
                - **Items** (Type: integer):
              - **Text** (Type: string):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **X** (Type: number):
                - **Y** (Type: number):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
                - **R** (Type: number):
          - **Single** (Type: array):
            - **Items** (Type: object):
              - **Reln** (Type: array):
                - **Items** (Type: integer):
              - **Text** (Type: string):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
                - **R** (Type: number):
                - **X** (Type: number):
                - **Y** (Type: number):
          - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
            - **Y** (Type: number):
            - **Z** (Type: number):
            - **Lat** (Type: number):
            - **Lon** (Type: number):
            - **R** (Type: number):
            - **X** (Type: number):
          - **Chapter** (Type: string):
          - **Common** (Type: array):
            - **Items** (Type: object):
              - **XYZ**: The coordinates of the node, relative to a search in 3d semantic spacetime (Type: object):
                - **Y** (Type: number):
                - **Z** (Type: number):
                - **Lat** (Type: number):
                - **Lon** (Type: number):
                - **R** (Type: number):
                - **X** (Type: number):
              - **Reln** (Type: array):
                - **Items** (Type: integer):
              - **Text** (Type: string):
      - **Option 7** (Type: array):
        - **Items**: Spceific information explaining a single type of arrow (Type: object):
          - **Long** (Type: string):
          - **Short** (Type: string):
          - **ASTtype** (Type: integer):
          - **ArrPtr** (Type: integer):
          - **ISTtype** (Type: integer):
          - **InvL** (Type: string):
          - **InvPtr** (Type: integer):
          - **InvS** (Type: string):
`

// NewN4LqueryMCPTool creates the MCP Tool instance for N4Lquery

func NewN4LqueryMCPTool() mcp.Tool {

	schema, err := json.Marshal(N4LqueryInputSchema)

	if err != nil {
		fmt.Printf("N4Lquery: failed to marshal help text: %v", err)
		os.Exit(-1)
	}
	
	return mcp.NewToolWithRawSchema(
		"N4Lquery",
		"Interpret search command",
		schema,
	)
}

// N4LqueryHandler is the handler function for the N4Lquery tool.
// This function is automatically generated. Users should implement the actual
// logic within this function body to integrate with backend APIs.
// You can generate types, http client and helpers for parsing request params to facilitate the implementation.

func N4LqueryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	/* type CallToolRequest struct {
                  JSONRPC string `json:"jsonrpc"`
                  ID      string `json:"id"`
                  Method  string `json:"method"` // Always "tools/call"
                  Params  struct {
                  Name      string                 `json:"name"`
                  Arguments map[string]interface{} `json:"arguments"`
                  } `json:"params"`
          }*/

	var mcp_search_command string
	
	if args, ok := request.Params.Arguments.(map[string]any); ok {
		//mcp_search_command = args["command_request"].(string)
		if body, ok := args["body"].(map[string]any); ok {
			if name, ok := body["name"].(string); ok {
				mcp_search_command = name
			}
		}
		fmt.Println("DEBUG ARG string:",request.Params.Name,"ARGS",mcp_search_command)
	}
	
	// Make HTTP calls or interact with services as needed.
	// Return an *mcp.CallToolResult with the response payload, or an error.
	// We need to submit a simple form to http_server

	formdata := url.Values{
		"name": { mcp_search_command },
	}

	// Reroute query to the SST server

	uri := "https://127.0.0.1:8443/searchN4L"

	var body []byte
	
	fmt.Println("Test the certificate",uri,formdata)
	
	resp, err := http.PostForm(uri, formdata)
	
	if err != nil {
		fmt.Printf("POST: Unable to forward request: %s\n", "N4Lquery")
		body = SelfSignedForm(uri,mcp_search_command,formdata)
		if body == nil {
			return nil, fmt.Errorf("%s upstream unreachable", "N4Lquery")
		}
	} else {
		defer resp.Body.Close()
		body, _ = io.ReadAll(resp.Body)
	}
	
	/* type CallToolResult struct {
                 Content []Content `json:"content"`
                 IsError bool      `json:"isError,omitempty"`
        } */
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(string(body))},
	}, nil
}


// *********************************************************************

func SelfSignedForm(uri,query string,formdata url.Values) []byte {
	
	// curl -Iv https://127.0.0.1:8443 --cacert cert.pem

	fmt.Println("Reading a self-signed certificate",SELF_SIGNED_CERT)
	caCert, err := os.ReadFile(SELF_SIGNED_CERT)
	
	if err != nil {
		fmt.Println("Couldn't load server's self-signed certificate file",SELF_SIGNED_CERT,err)
		return nil
	}
	
	// 2. Create a CertPool and add the CA certificate

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	
	// 3. Configure TLS with the custom CertPool
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	
	// 4. Create an HTTP client with the custom TLS configuration
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	
	fmt.Println("Try to connect to form interface...",uri,formdata)
	
	resp, err2 := client.PostForm(uri, formdata)
	
	if err2 != nil {
		fmt.Printf("POST: Unable to forward request: %s\n", "N4Lquery")
		return nil
	}
	
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)

	fmt.Println("Succeeded, sending response...")
	return body
}
