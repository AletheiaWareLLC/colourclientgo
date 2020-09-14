/*
 * Copyright 2020 Aletheia Ware LLC
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
	"flag"
	"fmt"
	"github.com/AletheiaWareLLC/bcclientgo"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/colourclientgo"
	"github.com/AletheiaWareLLC/colourgo"
	"io"
	"log"
	"os"
)

var peer = flag.String("peer", "", "Colour peer")

func main() {
	// Parse command line flags
	flag.Parse()

	// Set log flags
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	client := &colourclientgo.Client{
		BCClient: bcclientgo.BCClient{},
	}
	client.SetPeers(bcgo.SplitRemoveEmpty(*peer, ",")...)

	args := flag.Args()

	if len(args) > 0 {
		switch args[0] {
		case "init":
			PrintLegalese(os.Stdout)
			node, err := client.Init(&bcgo.PrintingMiningListener{Output: os.Stdout})
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("Initialized")
			if err := bcclientgo.PrintNode(os.Stdout, node); err != nil {
				log.Println(err)
				return
			}
		case "list":
			node, err := client.GetNode()
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("Canvases:")
			count := 0
			if err := client.List(node, func(entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error {
				count += 1
				return PrintCanvasShort(os.Stdout, entry, canvas)
			}); err != nil {
				log.Println(err)
				return
			}
			log.Println(count, "canvases")
		case "show":
			if len(args) > 1 {
				node, err := client.GetNode()
				if err != nil {
					log.Println(err)
					return
				}
				recordHash, err := base64.RawURLEncoding.DecodeString(args[1])
				if err != nil {
					log.Println(err)
					return
				}
				if err := client.Show(node, recordHash, func(entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error {
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
				node, err := client.GetNode()
				if err != nil {
					log.Println(err)
					return
				}
				log.Println("Canvases:")
				count := 0
				if client.ShowAll(node, args[1], func(entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error {
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
	// TODO fmt.Fprintln(output, "\tcolour render [hash] - displays a rendering of the canvas with given hash")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "\tcolour purchase [canvas] [location] [colour] [price] - creates a new purchase on the given canvas at the given location for the given colour with the given price")
	fmt.Fprintln(output, "\tcolour vote [canvas] [location] [colour] - creates a new vote on the given canvas at the given location for the given colour")
}

func PrintLegalese(output io.Writer) {
	fmt.Fprintln(output, "Colour Legalese:")
	fmt.Fprintln(output, "Colour is made available by Aletheia Ware LLC [https://aletheiaware.com] under the Terms of Service [https://aletheiaware.com/terms-of-service.html] and Privacy Policy [https://aletheiaware.com/privacy-policy.html].")
	fmt.Fprintln(output, "This beta version of Colour is made available under the Beta Test Agreement [https://aletheiaware.com/colour-beta-test-agreement.html].")
	fmt.Fprintln(output, "By continuing to use this software you agree to the Terms of Service, Privacy Policy, and Beta Test Agreement.")
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
