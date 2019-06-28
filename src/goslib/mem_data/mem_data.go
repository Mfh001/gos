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
package mem_data

import (
	"github.com/mafei198/gos/goslib/gen_server"
)

type Set map[string]interface{}
type Sets map[string]Set
type NamespaceSet map[string]Sets

type MemData struct {
	set NamespaceSet
}

var setServer *gen_server.GenServer

func Start() {
	var err error
	if setServer, err = gen_server.New(&MemData{set: NamespaceSet{}}); err != nil {
		panic(err)
	}
}

type SAddParam struct {
	namespace string
	key       string
	setId     string
	setValue  interface{}
}

func SAdd(namespace, key, setId string, value interface{}) {
	setServer.Cast(&SAddParam{namespace, key, setId, value})
}

type SRemParam struct {
	namespace string
	key       string
	setId     string
}

func SRem(namespace, key, setId string) {
	setServer.Cast(&SRemParam{namespace, key, setId})
}

type SCardParam struct {
	namespace string
	key       string
}

func SCard(namespace, key string) int32 {
	v, err := setServer.Call(&SCardParam{namespace, key})
	if err != nil {
		return 0
	}
	return v.(int32)
}

type SMembersParam struct {
	namespace string
	key       string
}

func SMembers(namespace, key string) Set {
	v, err := setServer.Call(&SMembersParam{namespace, key})
	if err != nil {
		return Set{}
	}
	return v.(Set)
}

type DelSetParam struct {
	namespace string
	key       string
}

func DelSet(namespace, key string) {
	setServer.Cast(&DelSetParam{namespace, key})
}

func (self *MemData) Init(args []interface{}) (err error) {
	return nil
}

func (self *MemData) HandleCall(req *gen_server.Request) (interface{}, error) {
	switch params := req.Msg.(type) {
	case *SCardParam:
		set := getSet(self.set, params.namespace, params.key)
		return len(set), nil
	case *SMembersParam:
		set := getSet(self.set, params.namespace, params.key)
		return set, nil
	}
	return nil, nil
}

func (self *MemData) HandleCast(req *gen_server.Request) {
	switch params := req.Msg.(type) {
	case *SAddParam:
		set := getSet(self.set, params.namespace, params.key)
		set[params.setId] = params.setValue
		break
	case *SRemParam:
		set := getSet(self.set, params.namespace, params.key)
		delete(set, params.setId)
		break
	case *DelSetParam:
		delSet(self.set, params.namespace, params.key)
	}
}

type TaskParams struct{}

func (self *MemData) handleTask(params *TaskParams) {
}

func (self *MemData) Terminate(reason string) (err error) {
	return nil
}

func getSet(namespaceSet NamespaceSet, namespace, key string) Set {
	sets, ok := namespaceSet[namespace]
	if !ok {
		sets = Sets{}
		namespaceSet[namespace] = sets
	}
	set, ok := sets[key]
	if !ok {
		set = Set{}
		sets[key] = set
	}
	return set
}

func delSet(namespaceSet NamespaceSet, namespace, key string) {
	if sets, ok := namespaceSet[namespace]; ok {
		if _, ok := sets[key]; ok {
			delete(sets, key)
		}
	}
}
