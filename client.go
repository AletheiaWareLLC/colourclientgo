/*
 * Copyright 2019 Aletheia Ware LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/base64"
	"fmt"
	"github.com/AletheiaWareLLC/aliasgo"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/colourgo"
	"io"
	"log"
	"os"
)

type CanvasCallback func(entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error

type Client struct {
	Root    string
	Cache   bcgo.Cache
	Network bcgo.Network
}

func (c *Client) Init() (*bcgo.Node, error) {
	// Add Colour host to peers
	if err := bcgo.AddPeer(c.Root, colourgo.GetColourHost()); err != nil {
		return nil, err
	}

	// Add BC host to peers
	if err := bcgo.AddPeer(c.Root, bcgo.GetBCHost()); err != nil {
		return nil, err
	}

	node, err := bcgo.GetNode(c.Root, c.Cache, c.Network)
	if err != nil {
		return nil, err
	}

	// Open Alias Channel
	aliases := aliasgo.OpenAliasChannel()
	if err := bcgo.LoadHead(aliases, c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	if err := bcgo.Pull(aliases, c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	if err := aliases.UniqueAlias(c.Cache, c.Network, node.Alias); err != nil {
		return nil, err
	}
	if err := aliasgo.RegisterAlias(bcgo.GetBCWebsite(), node.Alias, node.Key); err != nil {
		log.Println("Could not register alias remotely: ", err)
		log.Println("Registering locally")
		// Create record
		record, err := aliasgo.CreateSignedAliasRecord(node.Alias, node.Key)
		if err != nil {
			return nil, err
		}

		// Write record to cache
		reference, err := bcgo.WriteRecord(aliasgo.ALIAS, node.Cache, record)
		if err != nil {
			return nil, err
		}
		log.Println("Wrote Record", base64.RawURLEncoding.EncodeToString(reference.RecordHash))

		// Mine record into blockchain
		hash, _, err := node.Mine(aliases, &bcgo.PrintingMiningListener{os.Stdout})
		if err != nil {
			return nil, err
		}
		log.Println("Mined Alias", base64.RawURLEncoding.EncodeToString(hash))

		// Push update to peers
		if err := bcgo.Push(aliases, node.Cache, node.Network); err != nil {
			return nil, err
		}
	}
	return node, nil
}

func (c *Client) List(node *bcgo.Node, callback CanvasCallback) error {
	canvases := colourgo.OpenCanvasChannel()
	if err := bcgo.LoadHead(canvases, c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	if err := bcgo.Pull(canvases, c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	return colourgo.GetCanvas(canvases, c.Cache, c.Network, node.Alias, node.Key, nil, func(entry *bcgo.BlockEntry, key []byte, canvas *colourgo.Canvas) error {
		return callback(entry, canvas)
	})
}

func (c *Client) Show(node *bcgo.Node, recordHash []byte, callback CanvasCallback) error {
	canvases := colourgo.OpenCanvasChannel()
	if err := bcgo.LoadHead(canvases, c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	if err := bcgo.Pull(canvases, c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	return colourgo.GetCanvas(canvases, c.Cache, c.Network, node.Alias, node.Key, recordHash, func(entry *bcgo.BlockEntry, key []byte, canvas *colourgo.Canvas) error {
		return callback(entry, canvas)
	})
}

func (c *Client) ShowAll(node *bcgo.Node, mode string, callback CanvasCallback) error {
	canvases := colourgo.OpenCanvasChannel()
	if err := bcgo.LoadHead(canvases, c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	if err := bcgo.Pull(canvases, c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	return colourgo.GetCanvas(canvases, c.Cache, c.Network, node.Alias, node.Key, nil, func(entry *bcgo.BlockEntry, key []byte, canvas *colourgo.Canvas) error {
		// TODO check this comparison works
		if canvas.Mode.String() == mode {
			return callback(entry, canvas)
		}
		return nil
	})
}

func (c *Client) Handle(args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "init":
			node, err := c.Init()
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("Initialized")
			log.Println(node.Alias)
			publicKeyBytes, err := bcgo.RSAPublicKeyToPKIXBytes(&node.Key.PublicKey)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(base64.RawURLEncoding.EncodeToString(publicKeyBytes))
		case "list":
			node, err := bcgo.GetNode(c.Root, c.Cache, c.Network)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("Canvases:")
			count := 0
			if err := c.List(node, func(entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error {
				count += 1
				return PrintCanvasShort(os.Stdout, entry, canvas)
			}); err != nil {
				log.Println(err)
				return
			}
			log.Println(count, "canvases")
		case "show":
			if len(args) > 1 {
				node, err := bcgo.GetNode(c.Root, c.Cache, c.Network)
				if err != nil {
					log.Println(err)
					return
				}
				recordHash, err := base64.RawURLEncoding.DecodeString(args[1])
				if err != nil {
					log.Println(err)
					return
				}
				if err := c.Show(node, recordHash, func(entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error {
					return PrintCanvasLong(os.Stdout, entry, canvas)
				}); err != nil {
					log.Println(err)
					return
				}
			} else {
				log.Println("show <canvas-hash>")
			}
		case "showall":
			if len(args) > 1 {
				node, err := bcgo.GetNode(c.Root, c.Cache, c.Network)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println("Canvases:")
				count := 0
				if c.ShowAll(node, args[1], func(entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error {
					count += 1
					return PrintCanvasShort(os.Stdout, entry, canvas)
				}); err != nil {
					log.Println(err)
					return
				}
				log.Println(count, "canvases")
			} else {
				log.Println("showall <mode>")
			}
		default:
			log.Println("Cannot handle", args[0])
		}
	} else {
		PrintUsage(os.Stdout)
	}
}

func PrintUsage(output io.Writer) {
	fmt.Fprintln(output, "Colour Usage:")
	fmt.Fprintln(output, "\tcolour - print usage")
	fmt.Fprintln(output, "\tcolour init - initializes environment, generates key pair, and registers alias")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "\tcolour list - displays all canvases")
	fmt.Fprintln(output, "\tcolour show [hash] - display metadata of canvas with given hash")
	fmt.Fprintln(output, "\tcolour showall [mode] - display metadata of all canvases with given mode")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "\tcolour purchase [canvas] [location] [colour] [price] - posts a new record to Aletheia Ware's Purchasing Market")
	fmt.Fprintln(output, "\tcolour vote [canvas] [location] [colour] - posts a new record to Aletheia Ware's Voting Platform")
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get root directory
	rootDir, err := bcgo.GetRootDirectory()
	if err != nil {
		log.Fatal("Could not get root directory:", err)
	}
	log.Println("Root Directory:", rootDir)

	// Get cache directory
	cacheDir, err := bcgo.GetCacheDirectory(rootDir)
	if err != nil {
		log.Fatal("Could not get cache directory:", err)
	}
	log.Println("Cache Directory:", cacheDir)

	// Create file cache
	cache, err := bcgo.NewFileCache(cacheDir)
	if err != nil {
		log.Fatal("Could not create file cache:", err)
	}

	// Get peers
	peers, err := bcgo.GetPeers(rootDir)
	if err != nil {
		log.Fatal("Could not get network peers:", err)
	}

	// Create network of peers
	network := &bcgo.TcpNetwork{peers}

	client := &Client{
		Root:    rootDir,
		Cache:   cache,
		Network: network,
	}

	client.Handle(os.Args[1:])
}

func PrintCanvasShort(output io.Writer, entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error {
	hash := base64.RawURLEncoding.EncodeToString(entry.RecordHash)
	timestamp := bcgo.TimestampToString(entry.Record.Timestamp)
	fmt.Fprintf(output, "%s %s %s %d %d %d %s", hash, timestamp, canvas.Name, canvas.Width, canvas.Height, canvas.Depth, canvas.Mode)
	return nil
}

func PrintCanvasLong(output io.Writer, entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error {
	fmt.Fprintf(output, "Hash: %s", base64.RawURLEncoding.EncodeToString(entry.RecordHash))
	fmt.Fprintf(output, "Timestamp: %s", bcgo.TimestampToString(entry.Record.Timestamp))
	fmt.Fprintf(output, "Name: %s", canvas.Name)
	fmt.Fprintf(output, "Width: %d", canvas.Width)
	fmt.Fprintf(output, "Height: %d", canvas.Height)
	fmt.Fprintf(output, "Depth: %d", canvas.Depth)
	fmt.Fprintf(output, "Mode: %s", canvas.Mode)
	fmt.Fprintf(output, "References: %d", len(entry.Record.Reference))
	for index, reference := range entry.Record.Reference {
		fmt.Fprintf(output, "\t%d: %s", index, base64.RawURLEncoding.EncodeToString(reference.RecordHash))
	}
	return nil
}
