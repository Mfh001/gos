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

package data

import (
	"github.com/mafei198/gos/goslib/gen/db"
	"github.com/mafei198/gos/goslib/player"
)

func User(ctx *player.Player) *db.User {
	user := ctx.Data.User
	if user.MaxMonsterLevel == 0 {
		user.MaxMonsterLevel = 1
	}
	return user
}

func Equips(ctx *player.Player) map[int32]*db.Equip {
	data := ctx.Data
	if data.Equips == nil {
		data.Equips = map[int32]*db.Equip{}
	}
	return data.Equips
}

func AutoIncr(ctx *player.Player) *db.AutoIncr {
	autoIncr := ctx.Data.AutoIncr
	if autoIncr == nil {
		autoIncr = &db.AutoIncr{
			BuildingIdx: 1,
		}
		ctx.Data.AutoIncr = autoIncr
	}
	return autoIncr
}

