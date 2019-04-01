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
