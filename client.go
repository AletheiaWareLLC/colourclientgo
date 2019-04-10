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
	//"bytes"
	//"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/AletheiaWareLLC/aliasgo"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/colourgo"
	"github.com/AletheiaWareLLC/financego"
	//"github.com/golang/protobuf/proto"
	"log"
	"os"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) > 1 {
		// Handle Arguments
		switch os.Args[1] {
		case "init":
			if err := bcgo.AddPeer(colourgo.GetColourHost()); err != nil {
				log.Println(err)
				return
			}
			if err := bcgo.AddPeer(bcgo.GetBCHost()); err != nil {
				log.Println(err)
				return
			}
			aliases, err := aliasgo.OpenAliasChannel()
			if err != nil {
				log.Println(err)
				return
			}
			node, err := bcgo.GetNode()
			if err != nil {
				log.Println(err)
				return
			}
			alias, err := aliasgo.RegisterAlias(aliases, colourgo.GetColourWebsite(), node.Alias, node.Key)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(alias)
			publicKeyBytes, err := bcgo.RSAPublicKeyToPKIXBytes(&node.Key.PublicKey)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(base64.RawURLEncoding.EncodeToString(publicKeyBytes))
			log.Println("Initialized")
		case "list":
			node, err := bcgo.GetNode()
			if err != nil {
				log.Println(err)
				return
			}
			canvases, err := bcgo.OpenAndSyncChannel(colourgo.COLOUR_PREFIX_CANVAS + GetYear())
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("Canvases:")
			count := 0
			// List canvases
			err = colourgo.GetCanvas(canvases, node.Alias, node.Key, nil, func(entry *bcgo.BlockEntry, key []byte, canvas *colourgo.Canvas) error {
				count = count + 1
				return ShowCanvasShort(entry, canvas)
			})
			if err != nil {
				log.Println(err)
				return
			}
			log.Println(count, "canvases")
		case "show":
			// Show canvas with given hash
			if len(os.Args) > 2 {
				recordHash, err := base64.RawURLEncoding.DecodeString(os.Args[2])
				if err != nil {
					log.Println(err)
					return
				}
				node, err := bcgo.GetNode()
				if err != nil {
					log.Println(err)
					return
				}
				canvases, err := bcgo.OpenAndSyncChannel(colourgo.COLOUR_PREFIX_CANVAS + GetYear())
				if err != nil {
					log.Println(err)
					return
				}
				err = colourgo.GetCanvas(canvases, node.Alias, node.Key, recordHash, func(entry *bcgo.BlockEntry, key []byte, canvas *colourgo.Canvas) error {
					return ShowCanvasLong(entry, canvas)
				})
				if err != nil {
					log.Println(err)
					return
				}
			} else {
				log.Println("show <canvas-hash>")
			}
		case "showall":
			// Show all canvases with given mode
			if len(os.Args) > 2 {
				node, err := bcgo.GetNode()
				if err != nil {
					log.Println(err)
					return
				}
				canvases, err := bcgo.OpenAndSyncChannel(colourgo.COLOUR_PREFIX_CANVAS + GetYear())
				if err != nil {
					log.Println(err)
					return
				}
				log.Println("Canvases:")
				count := 0
				err = colourgo.GetCanvas(canvases, node.Alias, node.Key, nil, func(entry *bcgo.BlockEntry, key []byte, canvas *colourgo.Canvas) error {
					if canvas.Mode.String() == os.Args[2] {
						count = count + 1
						return ShowCanvasShort(entry, canvas)
					}
					return nil
				})
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(count, "canvases")
			} else {
				log.Println("showall <mode>")
			}
		case "customer":
			node, err := bcgo.GetNode()
			if err != nil {
				log.Println(err)
				return
			}
			customers, err := financego.OpenCustomerChannel()
			if err != nil {
				log.Println(err)
				return
			}
			customer, err := financego.GetCustomerSync(customers, node.Alias, node.Key, node.Alias)
			if err != nil {
				publicKeyBytes, err := bcgo.RSAPublicKeyToPKIXBytes(&node.Key.PublicKey)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(err)
				log.Println("To register as a Colour customer, visit", colourgo.GetColourWebsite()+"/colour-register?alias="+node.Alias)
				log.Println("and enter your alias, email, payment info, and public key:")
				log.Println(base64.RawURLEncoding.EncodeToString(publicKeyBytes))
				return
			}
			log.Println(customer)
		case "subscription":
			node, err := bcgo.GetNode()
			if err != nil {
				log.Println(err)
				return
			}
			subscriptions, err := financego.OpenSubscriptionChannel()
			if err != nil {
				log.Println(err)
				return
			}
			subscription, err := financego.GetSubscriptionSync(subscriptions, node.Alias, node.Key, node.Alias)
			if err != nil {
				log.Println(err)
				log.Println("To subscribe for purchase market and voting platform, visit", colourgo.GetColourWebsite()+"/colour-subscribe?alias="+node.Alias)
				log.Println("and enter your alias, and customer ID")
			} else {
				log.Println(subscription)
			}
		case "stripe-webhook":
			// TODO
		default:
			PrintUsage()
		}
	} else {
		PrintUsage()
	}
}

func PrintUsage() {
	log.Println("Colour Usage:")
	log.Println("\tcolour - display usage")
	log.Println("\tcolour init - initializes environment, generates key pair, and registers alias")
	log.Println("\tcolour list - initializes environment, generates key pair, and registers alias")
	log.Println("\tcolour show - initializes environment, generates key pair, and registers alias")
	log.Println("\tcolour showall - initializes environment, generates key pair, and registers alias")
	log.Println("\tcolour list - displays all canvases")
	log.Println("\tcolour show [hash] - display metadata of canvas with given hash")
	log.Println("\tcolour showall [type] - display metadata of all canvases with given mode")

	log.Println("\tcolour customer - display Stripe customer information")
	log.Println("\tcolour subscription - display String subscription information")
	log.Println("\tcolour purchase [canvas] [location] [colour] [price] - posts a new record to Aletheia Ware's Purchasing Market")
	log.Println("\tcolour vote [canvas] [location] [colour] - posts a new record to Aletheia Ware's Voting Platform")

	log.Println("BC Usage:")
	log.Println("\tbc sync [channel] - synchronizes cache for given channel")
	log.Println("\tbc head [channel] - display head of given channel")
	log.Println("\tbc block [channel] [hash] - display block with given hash on given channel")
	log.Println("\tbc record [channel] [hash] - display record with given hash on given channel")

	log.Println("\tbc alias [alias] - display public key for alias")
	log.Println("\tbc node - display registered alias and public key")

	log.Println("\tbc import-keys [alias] [access-code] - imports the alias and keypair from BC server")
	log.Println("\tbc export-keys [alias] - generates a new access code and exports the alias and keypair to BC server")

	log.Println("\tbc cache - display location of cache")
	log.Println("\tbc keystore - display location of keystore")
	log.Println("\tbc peers - display list of peers")
	log.Println("\tbc add-peer [host] - adds the given host to the list of peers")

	log.Println("\tbc random - generate a random number")
}

func GetYear() string {
	return fmt.Sprintf("%d", time.Now().Year())
}

func ShowCanvasShort(entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error {
	hash := base64.RawURLEncoding.EncodeToString(entry.RecordHash)
	timestamp := bcgo.TimestampToString(entry.Record.Timestamp)
	log.Println(hash, timestamp, canvas.Name, canvas.Width, canvas.Height, canvas.Depth, canvas.Mode)
	return nil
}

func ShowCanvasLong(entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error {
	hash := base64.RawURLEncoding.EncodeToString(entry.RecordHash)
	timestamp := bcgo.TimestampToString(entry.Record.Timestamp)
	log.Println("Hash:", hash)
	log.Println("Timestamp:", timestamp)
	log.Println("Name:", canvas.Name)
	log.Println("Width:", canvas.Width)
	log.Println("Height:", canvas.Height)
	log.Println("Depth:", canvas.Depth)
	log.Println("Mode:", canvas.Mode)
	log.Println("References:", len(entry.Record.Reference))
	for index, reference := range entry.Record.Reference {
		hash := base64.RawURLEncoding.EncodeToString(reference.RecordHash)
		log.Println("\t", index, hash)
	}
	return nil
}
