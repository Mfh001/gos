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

class Cmds
  # Gen dockerfile
  def self.gen_dockerfile(docker_filename, app_name)
    File.open(docker_filename, "w") do |io|
      io.write(%Q{\
  FROM alpine
  ADD ./src/#{app_name}/bin/#{app_name} /
  CMD ["/#{app_name}"]
  })
    end
  end

  # Build docker image
  def self.build_image(image_name, docker_filename)
    $stdout.puts `docker build -t #{image_name}:latest -f #{docker_filename} ../`
  end

  # Export docker image
  def self.export_image(image_name)
    $stdout.puts `docker image save -o ./images/#{image_name}.tar #{image_name}`
  end

  def self.load_image(image_name)
    puts "Load image: #{image_name}.tar"
    $stdout.puts `docker load < images/#{image_name}.tar`
  end

  def self.build_apps
    $stdout.puts `docker run --rm -v "$PWD/../":/usr/src/gos -w /usr/src/gos golang:alpine sh build_gos.sh`
  end

  def self.run_app(image_name, port_mapping)
    puts "Run docker: #{image_name} -> #{port_mapping}"
    $stdout.puts `docker run -d #{port_mapping} #{image_name}`
  end
end
