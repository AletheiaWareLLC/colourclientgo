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
)

type CanvasCallback func(entry *bcgo.BlockEntry, canvas *colourgo.Canvas) error

type ColourClient struct {
	bcclientgo.BCClient
}

func (c *ColourClient) Init(listener bcgo.MiningListener) (*bcgo.Node, error) {
	// Add Colour host to peers
	if err := bcgo.AddPeer(c.Root, colourgo.GetColourHost()); err != nil {
		return nil, err
	}

	// Add BC host to peers
	if err := bcgo.AddPeer(c.Root, bcgo.GetBCHost()); err != nil {
		return nil, err
	}

	return c.BCClient.Init(listener)
}

func (c *ColourClient) List(node *bcgo.Node, callback CanvasCallback) error {
	name := colourgo.GetCanvasChannelName()
	canvases := node.GetOrOpenChannel(name, colourgo.OpenCanvasChannel)
	return colourgo.GetCanvas(canvases, c.Cache, c.Network, node.Alias, node.Key, nil, func(entry *bcgo.BlockEntry, key []byte, canvas *colourgo.Canvas) error {
		return callback(entry, canvas)
	})
}

func (c *ColourClient) Show(node *bcgo.Node, recordHash []byte, callback CanvasCallback) error {
	name := colourgo.GetCanvasChannelName()
	canvases := node.GetOrOpenChannel(name, colourgo.OpenCanvasChannel)
	return colourgo.GetCanvas(canvases, c.Cache, c.Network, node.Alias, node.Key, recordHash, func(entry *bcgo.BlockEntry, key []byte, canvas *colourgo.Canvas) error {
		return callback(entry, canvas)
	})
}

func (c *ColourClient) ShowAll(node *bcgo.Node, mode string, callback CanvasCallback) error {
	name := colourgo.GetCanvasChannelName()
	canvases := node.GetOrOpenChannel(name, colourgo.OpenCanvasChannel)
	return colourgo.GetCanvas(canvases, c.Cache, c.Network, node.Alias, node.Key, nil, func(entry *bcgo.BlockEntry, key []byte, canvas *colourgo.Canvas) error {
		if canvas.Mode.String() == mode {
			return callback(entry, canvas)
		}
		return nil
	})
}
