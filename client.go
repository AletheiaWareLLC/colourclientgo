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

package colourclientgo

import (
	"github.com/AletheiaWareLLC/bcclientgo"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/colourgo"
	"log"
)

type CanvasCallback func(entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error

type Client struct {
	bcclientgo.BCClient
}

func (c *Client) List(node *bcgo.Node, callback CanvasCallback) error {
	canvases := colourgo.OpenCanvasChannel()
	if err := canvases.LoadHead(c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	if err := canvases.Pull(c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	return colourgo.GetCanvas(canvases, c.Cache, c.Network, node.Alias, node.Key, nil, func(entry *bcgo.BlockEntry, key []byte, canvas *colourgo.Canvas) error {
		return callback(entry, canvas)
	})
}

func (c *Client) Show(node *bcgo.Node, recordHash []byte, callback CanvasCallback) error {
	canvases := colourgo.OpenCanvasChannel()
	if err := canvases.LoadHead(c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	if err := canvases.Pull(c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	return colourgo.GetCanvas(canvases, c.Cache, c.Network, node.Alias, node.Key, recordHash, func(entry *bcgo.BlockEntry, key []byte, canvas *colourgo.Canvas) error {
		return callback(entry, canvas)
	})
}

func (c *Client) ShowAll(node *bcgo.Node, mode string, callback CanvasCallback) error {
	canvases := colourgo.OpenCanvasChannel()
	if err := canvases.LoadHead(c.Cache, c.Network); err != nil {
		log.Println(err)
	}
	if err := canvases.Pull(c.Cache, c.Network); err != nil {
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
