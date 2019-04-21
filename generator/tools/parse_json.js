const JsonToGo = require("./json-to-go");
const fs = require("fs");
const folder = "./json_files";
const gen_dir = "../src/goslib/src/gen/gd/";
const gen_path = "../src/goslib/src/gen/gd/gd.go";

// First, checks if it isn't implemented yet.
if (!String.prototype.format) {
    String.prototype.format = function() {
        let args = arguments;
        return this.replace(/{(\d+)}/g, function(match, number) {
            return typeof args[number] != 'undefined'
                ? args[number]
                : match
                ;
        });
    };
}

let tpl = `type {0} struct {
	index map[int]int
	list  []*Config{0}
}

var {0}Ins = &{0}{index: map[int]int{}}

func (self *{0}) load(content string) {
	_ = json.UnmarshalFromString(content, &self.list)
	for i := 0; i < len(self.list); i++ {
		self.index[self.list[i].ID] = i
	}
}

func (self *{0}) GetItem(key int) *Config{0} {
	rwlock.RLock()
	defer rwlock.RUnlock()
	idx, ok := self.index[key]
	if !ok {
		return nil
	}
	return self.list[idx]
}
func (self *{0}) GetList() []*Config{0} {
	rwlock.RLock()
	defer rwlock.RUnlock()
	return self.list
}`;

let loadTpl = `
package gd

import (
	"github.com/json-iterator/go"
	"sync"
)

var rwlock = &sync.RWMutex{}
var json = jsoniter.ConfigCompatibleWithStandardLibrary

func LoadConfigs(data map[string]string) {
	rwlock.Lock()
	defer rwlock.Unlock()

    {0}
}
`;

fs.readdir(folder, (err, files) => {
    let loads = "";
    let package = {}
    files.forEach(file => {
        if (file.split('.').pop() === "json") {
            let name = file.split(".").shift();
            let content = fs.readFileSync(folder + "/" + file, { encoding: 'utf8' });
            let struct = JsonToGo(content, name).go;

            if (name === "Global") {
                let go_code = `
                    package gd
                    ${struct}
                    
                    var GlobalIns = &Global{}
                    func (self *Global) load(content string) {
                        _ = json.UnmarshalFromString(content, &GlobalIns)
                    }
                `;
                fs.writeFileSync(gen_dir + "/" + name + ".go", go_code, { encoding: 'utf8' });
                loads += `${name}Ins.load(data["${name}"])\n`
            } else {
                let relName = name.split("config")[1];
                let go_code = `
                    package gd
                    ${struct}
                    ${tpl.format(relName)}
                `;
                fs.writeFileSync(gen_dir + "/" + name + ".go", go_code, { encoding: 'utf8' });
                loads += `${relName}Ins.load(data["config${relName}"])\n`
            }
            package[name] = content
        }
    });
    gd_content = loadTpl.format(loads);
    fs.writeFileSync(gen_path, gd_content, { encoding: 'utf8' });
    fs.writeFileSync("./configData.json", JSON.stringify(package), { encoding: 'utf8' })
});
