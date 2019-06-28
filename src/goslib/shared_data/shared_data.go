/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package shared_data

import (
	"github.com/mafei198/gos/goslib/gen_server"
)

type LoadHandler func(table, key string) (interface{}, error)
type WriteHandler func(table, key string, rec interface{}) error
type DeleteHandler func(table, key string) error

type StartOptions struct {
	Name          string
	LoadHandler   LoadHandler
	CreateHandler WriteHandler
	UpdateHandler WriteHandler
	DeleteHandler DeleteHandler
}

/*
	Support realtime read write data, and persist data to redis async
*/

type SharedData struct {
	options   *StartOptions
	tables    map[string]map[string]interface{}
	loaded    map[string]map[string]bool
	persister *gen_server.GenServer
}

type FindParams struct {
	table string
	key   string
}

func (self *SharedData) Stop(reason string) error {
	return gen_server.Stop(self.options.Name, reason)
}

func (self *SharedData) Find(table, key string) interface{} {
	value, _ := gen_server.Call(self.options.Name, &FindParams{table, key})
	return value
}

type CreateParams struct {
	table string
	key   string
	rec   interface{}
}

func (self *SharedData) Create(table, key string, rec interface{}) {
	gen_server.Cast(self.options.Name, &CreateParams{table, key, rec})
}

type UpdateParams struct {
	table string
	key   string
	rec   interface{}
}

func (self *SharedData) Update(table, key string, rec interface{}) {
	gen_server.Cast(self.options.Name, &UpdateParams{table, key, rec})
}

type DeleteParams struct {
	table string
	key   string
}

func (self *SharedData) Delete(table, key string) {
	gen_server.Cast(self.options.Name, &DeleteParams{table, key})
}

func (self *SharedData) Init(args []interface{}) (err error) {
	self.tables = map[string]map[string]interface{}{}
	self.loaded = map[string]map[string]bool{}

	persister := &Persister{}
	if self.persister, err = gen_server.New(persister); err != nil {
		return err
	}
	persister.server = self.persister
	return nil
}

func (self *SharedData) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch params := req.Msg.(type) {
	case *FindParams:
		if err := self.ensureLoaded(params.table, params.key); err != nil {
			return nil, err
		}
		return self.tables[params.table][params.key], nil
	}
	return nil, nil
}

func (self *SharedData) HandleCast(req *gen_server.Request) {
	switch params := req.Msg.(type) {
	case *CreateParams:
		self.setCache(params.table, params.key, params.rec)
		self.asyncCreate(params)
	case *UpdateParams:
		self.setCache(params.table, params.key, params.rec)
		self.asyncUpdate(params)
	case *DeleteParams:
		if table := self.tables[params.table]; table != nil {
			delete(table, params.key)
			self.asyncDelete(params)
		}
	}
}

func (self *SharedData) Terminate(reason string) (err error) {
	return self.persister.Stop(reason)
}

func (self *SharedData) asyncCreate(params *CreateParams) {
	self.persister.Cast(&PersistParams{
		handler: self.options.CreateHandler,
		table:   params.table,
		key:     params.key,
		rec:     params.rec,
	})
}

func (self *SharedData) asyncUpdate(params *UpdateParams) {
	self.persister.Cast(&PersistParams{
		handler: self.options.UpdateHandler,
		table:   params.table,
		key:     params.key,
		rec:     params.rec,
	})
}

func (self *SharedData) asyncDelete(params *DeleteParams) {
	self.persister.Cast(&RemoveParams{
		handler: self.options.DeleteHandler,
		table:   params.table,
		key:     params.key,
	})
}

func (self *SharedData) ensureLoaded(tableName, key string) error {
	if cache := self.loaded[tableName]; cache != nil && cache[key] {
		return nil
	}
	value, err := self.options.LoadHandler(tableName, key)
	if err != nil {
		return err
	}
	self.setCache(tableName, key, value)
	return nil
}

func (self *SharedData) setCache(tableName, key string, rec interface{}) {
	cache := self.loaded[tableName]
	if cache == nil {
		cache = map[string]bool{}
		self.loaded[tableName] = cache
	}
	cache[key] = true

	table := self.tables[tableName]
	if table == nil {
		table = make(map[string]interface{})
		self.tables[tableName] = table
	}
	table[key] = rec
}
