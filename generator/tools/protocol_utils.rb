require "yaml"

## The MIT License (MIT)
##
## Copyright (c) 2018 SavinMax. All rights reserved.
##
## Permission is hereby granted, free of charge, to any person obtaining a copy
## of this software and associated documentation files (the "Software"), to deal
## in the Software without restriction, including without limitation the rights
## to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
## copies of the Software, and to permit persons to whom the Software is
## furnished to do so, subject to the following conditions:
##
## The above copyright notice and this permission notice shall be included in
## all copies or substantial portions of the Software.
##
## THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
## IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
## FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
## AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
## LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
## OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
## THE SOFTWARE.

class Utils
  def self.load_protos
    defines = []
    game_msgs = []
    room_msgs = []
    Dir.glob("./config/protocol/base/*.yaml").each do |source|
      defines += YAML.load_file(source)
    end
    Dir.glob("./config/protocol/game/*.yaml").each do |source|
      defines += YAML.load_file(source)
      game_msgs += YAML.load_file(source)
    end
    Dir.glob("./config/protocol/room/*.yaml").each do |source|
      defines += YAML.load_file(source)
      room_msgs += YAML.load_file(source)
    end
    return {defines: defines, game_msgs: game_msgs, room_msgs: room_msgs}
  end
end
